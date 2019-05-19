package service

import (
	"os"
	"errors"
	"context"
	"net/url"
	"math/big"

	ec "galaxyotc/ethwallet/config"
	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/ethereum_wallet"

	"galaxyotc/common/log"
	"galaxyotc/ethwallet/wallet"
	"galaxyotc/ethwallet/boltdb"

	"github.com/spf13/viper"
	"github.com/nats-io/go-nats"
)

type Service struct {
	rinkebyWallet *wallet.EthereumWallet
}

// 创建新服务
func NewService() *Service {
	u := viper.GetString("ethereum_wallet_service.url")
	log.Debug("url:", u)

	clientApi, _ := url.Parse(u)
	ethereumDS, err := boltdb.Create("rinkeby")
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	ethCfg := ec.CoinConfig{
		CoinType:  wi.Ethereum,
		FeeAPI:    url.URL{},
		LowFee:    *big.NewInt(4 * 1000 * 1000 * 1000), // 0.3Gwei
		MediumFee: *big.NewInt(7 * 1000 * 1000 * 1000),
		HighFee:   *big.NewInt(30 * 1000 * 1000 * 1000),
		MaxFee:    *big.NewInt(150 * 1000 * 1000 * 1000),
		ClientAPI: *clientApi,
		DB:        ethereumDS,
	}

	mnemonic := "label pyramid flat spike course crystal humor throw rug frozen food comic"

	service := &Service{}

	service.rinkebyWallet, err = wallet.NewEthereumWallet(ethCfg, mnemonic)
	if err != nil {
		log.Error(err)
		return nil
	}

	service.rinkebyWallet.AddEthereumTransactionListener(EthereumCallback)

	height, hash := service.rinkebyWallet.ChainTip()
	log.Debug("NewService, height:", height, ", hash:", hash)

	return service
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewEthereumWalletServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName+"_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	s.rinkebyWallet.Start()
	return nil
}

// 服务关闭
func (s *Service) Close() {
	s.rinkebyWallet.Close()
}

// 获取充值地址
func (s *Service) Deposit(ctx context.Context, req pb.DepositReq) (pb.DepositResp, error) {
	log.Infof("Visiting Deposit, Request Params is %+v", req)
	var purpose wi.KeyPurpose
	switch req.Purpose {
	case 0:
		purpose = wi.EXTERNAL
	case 1:
		purpose = wi.INTERNAL
	default:
		return pb.DepositResp{}, errors.New("无效的参数")
	}

	address, err := s.rinkebyWallet.Deposit(purpose)
	if err != nil {
		log.Errorf("server-Deposit-Error: %s", err.Error())
		return pb.DepositResp{}, err
	}
	return pb.DepositResp{Address: address}, nil
}

// 提现申请：发送ETH给接收者
func (s *Service) EtherWithdraw(ctx context.Context, req pb.EtherWithdrawReq) (pb.EtherWithdrawResp, error) {
	log.Infof("Visiting EtherWithdraw, Request Params is %+v", req)
	Value, ok := new(big.Int).SetString(req.Value, 10)
	if !ok {
		return pb.EtherWithdrawResp{}, errors.New("无效的提币金额")
	}

	txid, err := s.rinkebyWallet.EtherWithdraw(req.To, Value)
	if err != nil {
		log.Errorf("server-EtherWithdraw-Error: %s", err.Error())
		return pb.EtherWithdrawResp{}, err
	}
	return pb.EtherWithdrawResp{Txid: txid}, nil
}

// 提现申请：发送ERC20代币给接收者
func (s *Service) TokenWithdraw(ctx context.Context, req pb.TokenWithdrawReq) (pb.TokenWithdrawResp, error) {
	log.Infof("Visiting TokenWithdraw, Request Params is %+v", req)
	Value, ok := new(big.Int).SetString(req.Value, 10)
	if !ok {
		return pb.TokenWithdrawResp{}, errors.New("无效的提币金额")
	}

	txid, err := s.rinkebyWallet.TokenWithdraw(req.TokenAddress, req.To, Value)
	if err != nil {
		log.Errorf("server-TokenWithdraw-Error: %s", err.Error())
		return pb.TokenWithdrawResp{}, err
	}

	return pb.TokenWithdrawResp{Txid: txid}, err
}
