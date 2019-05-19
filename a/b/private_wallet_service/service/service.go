package service

import (
	"os"
	"time"
	"errors"
	"context"
	"net/url"
	"math/big"

	"galaxyotc/ethwallet/wallet"
	"galaxyotc/ethwallet/boltdb"

	ec "galaxyotc/ethwallet/config"
	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/private_wallet"
	ethCommon "github.com/ethereum/go-ethereum/common"

	"galaxyotc/common/log"

	"github.com/spf13/viper"
	"github.com/nats-io/go-nats"
	"galaxyotc/gc_services/private_wallet_service/api"
	"github.com/golang/protobuf/proto"
)

var (
	toBuyer int32 = 0
	toSeller int32 = 1
)

// 发送操作错误回调
func sendTokenErrorCallback(isDeposit bool, tokenAddress, whoAddress string) {
	// 交易信息
	transactionCallback := &pb.TransactionErrorCallback{
		IsDeposit: *proto.Bool(isDeposit),
		TokenAddress: *proto.String(tokenAddress),
		WhoAddress: *proto.String(whoAddress),
	}

	transactionByte, err := proto.Marshal(transactionCallback)
	if err != nil {
		log.Errorf("server-sendErrorCallback-Error: %s", err.Error())
	}
	if _, err := api.AccountApi.PrivateErrorCallback(transactionByte); err != nil  {
		log.Errorf("server-sendErrorCallback-Error: %s", err.Error())
	}
}

type Service struct {
	privateWallet *wallet.EthereumWallet
}

// 创建新服务
func NewService() *Service {
	clientApi, _ := url.Parse(viper.GetString("private_wallet_service.url"))
	ethereumDS, err := boltdb.Create("private")
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
	ethCfg := ec.CoinConfig{
		CoinType:  wi.Ethereum,
		FeeAPI:    url.URL{},
		LowFee:    *big.NewInt(4 * 1000*1000*1000), // 0.3Gwei
		MediumFee: *big.NewInt(7 * 1000*1000*1000),
		HighFee:   *big.NewInt(30 * 1000*1000*1000),
		MaxFee:    *big.NewInt(150 * 1000*1000*1000),
		ClientAPI: *clientApi,
		DB:        ethereumDS,
	}

	mnemonic:="label pyramid flat spike course crystal humor throw rug frozen food comic"

	privateWallet, _ := wallet.NewEthereumWallet(ethCfg, mnemonic)
	service := &Service{
		privateWallet: privateWallet,
	}
	return service
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewPrivateWalletServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	//s.privateWallet.Start()
	return nil
}

// 服务关闭
func (s *Service) Close() {
	s.privateWallet.Close()
}

// 区块当前高度
func (s *Service) ChainTip(ctx context.Context, req pb.ChainTipReq) (pb.ChainTipResp, error) {
	log.Infof("Visiting ChainTip, Request Params is %+v", req)
	height, hash := s.privateWallet.ChainTip()
	return pb.ChainTipResp{Height: height, Hash: hash.String()}, nil
}

// 给指定地址转账
func (s *Service) Transfer(ctx context.Context, req pb.TransferReq) (pb.TransferResp, error) {
	log.Infof("Visiting Transfer, Request Params is %+v", req)
	valueWei, _ := new(big.Int).SetString(req.Value, 10)

	hash, err := s.privateWallet.Transfer(req.To, valueWei)
	if err != nil {
		log.Errorf("server-Transfer-Error: %s", err.Error())
		return pb.TransferResp{}, err
	}
	return pb.TransferResp{Hash: hash.String()}, nil
}

// 获取一个新地址
func (s *Service) NewAddress(ctx context.Context, req pb.NewAddressReq) (pb.AddressResp, error) {
	log.Infof("Visiting NewAddress, Request Params is %+v", req)
	var purpose wi.KeyPurpose
	switch req.Purpose {
	case 0:
		purpose = wi.EXTERNAL
	case 1:
		purpose = wi.INTERNAL
	default:
		return pb.AddressResp{}, errors.New("无效的参数")
	}

	address := s.privateWallet.NewAddress(purpose)
	return pb.AddressResp{Address: address.String()}, nil
}

// 添加一个代币担保交易：即锁币
func (s *Service) AddTokenTransaction(ctx context.Context, req pb.TransactionReq) (pb.TransactionResp, error) {
	log.Infof("Visiting AddTokenTransaction, Request Params is %+v", req)
	// 唯一ID需要将订单流水号转为地址
	uniqueId := ethCommon.BytesToAddress([]byte(req.Sn)).String()
	// 将小数转换为位
	amountWei, _ := new(big.Int).SetString(req.Amount, 10)
	// 仲裁者地址
	moderatorAddress := s.privateWallet.OwnerAddress()
	// 智能合约地址
	escrowAddress := viper.GetString("private_wallet_service.escrow_address")

	txid, err := s.privateWallet.AddTokenTransaction(uniqueId, int(req.Threshold), time.Duration(req.Timeout), req.BuyerAddress, req.SellerAddress, moderatorAddress, amountWei, escrowAddress, req.TokenAddress)
	if err != nil {
		log.Errorf("server-AddTokenTransaction-Error: %s", err.Error())
		return pb.TransactionResp{}, err
	}
	return pb.TransactionResp{Txid: *txid}, nil
}

// 构造参数并释放一个代币担保交易:即放币
func (s *Service) ExecuteTransaction(ctx context.Context, req pb.TransactionReq) (pb.TransactionResp, error) {
	log.Infof("Visiting ExecuteTransaction, Request Params is %+v", req)
	// 唯一ID需要将订单流水号转为地址
	uniqueId := ethCommon.BytesToAddress([]byte(req.Sn)).String()
	// 将小数转换为位
	amountWei, _ := new(big.Int).SetString(req.Amount, 10)
	// 仲裁者地址
	moderatorAddress := s.privateWallet.OwnerAddress()
	// 智能合约地址
	escrowAddress :=  viper.GetString("private_wallet_service.escrow_address")

	//构造参数
	redeemScript := s.privateWallet.GenerateRedeemScript(uniqueId, int(req.Threshold), time.Duration(req.Timeout), req.BuyerAddress, req.SellerAddress, moderatorAddress, escrowAddress, req.TokenAddress)
	solidityHash := redeemScript.SoliditySHA3()
	// 因为构造哈希时，在内部进行了买方与卖方的交换，所以签名多签名的买方是实际交易的卖方
	signers := []*ethCommon.Address{&redeemScript.Moderator, &redeemScript.Buyer}
	payables := make(map[string]*big.Int)

	var toAddress string
	switch req.To {
	case toBuyer:
		// 将钱分配给买方，这里的卖方就是实际交易的买方
		toAddress = redeemScript.Seller.Hex()
	case toSeller:
		// 将钱退回给卖方
		payables[redeemScript.Buyer.Hex()] = amountWei
	}

	payables[toAddress] = amountWei

	txid, err := s.privateWallet.ExecuteTransaction(signers, payables, solidityHash)
	if err != nil {
		log.Errorf("server-ExecuteTransaction-Error: %s", err.Error())
		return pb.TransactionResp{}, err
	}
	return pb.TransactionResp{Txid: *txid}, nil
}

// 部署代币合约
func (s *Service) DeployToken(ctx context.Context, req pb.DeployTokenReq) (pb.AddressResp, error) {
	log.Infof("Visiting DeployToken, Request Params is %+v", req)
	address, err := s.privateWallet.DeployToken(req.Name, req.Symbol, uint8(req.Decimals))
	if err != nil {
		log.Errorf("server-DeployToken-Error: %s", err.Error())
		return pb.AddressResp{}, err
	}
	return pb.AddressResp{Address: address.String()}, nil
}

// 获取代币的余额
func (s *Service) GetTokenBalance(ctx context.Context, req pb.GetTokenBalanceReq) (pb.GetTokenBalanceResp, error) {
	log.Infof("Visiting GetTokenBalance, Request Params is %+v", req)
	balanceWei, err := s.privateWallet.GetTokenBalance(req.TokenAddress, req.WhoAddress)
	if err != nil {
		log.Errorf("server-GetTokenBalance-Error: %s", err.Error())
		return pb.GetTokenBalanceResp{}, err
	}
	return pb.GetTokenBalanceResp{Balance: balanceWei.String()}, nil
}

// 挖矿代币
func (s *Service) MintToken(ctx context.Context, req pb.TokenReq) (pb.TokenResp, error) {
	log.Infof("Visiting MintToken, Request Params is %+v", req)
	// 将小数转换为位
	amountWei, _ := new(big.Int).SetString(req.Amount, 10)

	err := s.privateWallet.MintToken(req.TokenAddress, req.WhoAddress, *amountWei)
	if err != nil {
		log.Errorf("server-MintToken-Error: %s", err.Error())
		sendTokenErrorCallback(true, req.TokenAddress, req.WhoAddress)
		return pb.TokenResp{}, err
	}
	return pb.TokenResp{}, nil
}

// 燃烧代币
func (s *Service) BurnToken(ctx context.Context, req pb.TokenReq) (pb.TokenResp, error) {
	log.Infof("Visiting BurnToken, Request Params is %+v", req)
	// 将小数转换为位
	amountWei, _ := new(big.Int).SetString(req.Amount, 10)

	err := s.privateWallet.BurnToken(req.TokenAddress, req.WhoAddress, *amountWei)
	if err != nil {
		log.Errorf("server-BurnToken-Error: %s", err.Error())
		sendTokenErrorCallback(false, req.TokenAddress, req.WhoAddress)
		return pb.TokenResp{}, err
	}
	return pb.TokenResp{}, nil
}