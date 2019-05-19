package service

import (
	"errors"
	"context"

	mc "galaxyotc/multiwallet/config"
	bwi "galaxyotc/btc-wallet-interface"
	pb "galaxyotc/common/proto/wallet/multi_wallet"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"galaxyotc/common/log"

	"galaxyotc/multiwallet"
	"galaxyotc/multiwallet/bitcoin"
	"github.com/spf13/viper"
	"github.com/nats-io/go-nats"
)

type Service struct {
	multiWallet *multiwallet.MultiWallet
}

// 创建新服务
func NewService() *Service {
	m := make(map[bwi.CoinType]bool)
	m[bwi.Bitcoin] = true
	m[bwi.BitcoinCash] = true
	m[bwi.Zcash] = false
	m[bwi.Litecoin] = true
	params := &chaincfg.MainNetParams
	if viper.GetBool("multi_wallet_service.test_net") {
		params = &chaincfg.TestNet3Params
	}
	cfg := mc.NewDefaultConfig(m, params)
	cfg.Mnemonic = "label pyramid flat spike course crystal humor throw rug frozen food comic"

	service := &Service{}
	service.multiWallet, _ = multiwallet.NewMultiWallet(cfg)

	for coinType, isTrue := range m {
		if isTrue {
			service.multiWallet.AddTransactionListener(coinType.CurrencyCode(), MultiCallback)
		}
	}

	return service
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewMultiWalletServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	s.multiWallet.Start()
	return nil
}

// 服务关闭
func (s *Service) Close() {
	s.multiWallet.Close()
}

// 区块当前高度
func (s *Service) ChainTip(ctx context.Context, req pb.ChainTipReq) (pb.ChainTipResp, error) {
	log.Infof("Visiting ChainTip, Request Params is %+v", req)
	wallet, err := s.multiWallet.WalletForCurrencyCode(req.Code)
	if err != nil {
		log.Errorf("server-ChainTip-Error: %s", err.Error())
		return pb.ChainTipResp{}, err
	}
	var (
		height uint32
		hash chainhash.Hash
	)
	if req.PropertyId != "0" {
		bitcoinWallet, ok := wallet.(*bitcoin.BitcoinWallet)
		if ok {
			height, hash = bitcoinWallet.ChainTip()
		}
	} else {
		height, hash = wallet.ChainTip()
	}
	return pb.ChainTipResp{Height: height, Hash: hash.String()}, nil
}

// 获取充值地址
func (s *Service) Deposit(ctx context.Context, req pb.DepositReq) (pb.DepositResp, error) {
	log.Infof("Visiting Deposit, Request Params is %+v", req)
	var purpose bwi.KeyPurpose
	switch req.Purpose {
	case 0:
		purpose = bwi.EXTERNAL
	case 1:
		purpose = bwi.INTERNAL
	default:
		return pb.DepositResp{}, errors.New("无效的参数")
	}

	address, err := s.multiWallet.Deposit(req.Code, req.PropertyId, purpose)
	if err != nil {
		log.Errorf("server-Deposit-Error: %s", err.Error())
		return pb.DepositResp{}, err
	}
	return pb.DepositResp{Address: address}, nil
}

// 提币
func (s *Service) Withdraw(ctx context.Context, req pb.WithdrawReq) (pb.WithdrawResp, error) {
	log.Infof("Visiting Withdraw, Request Params is %+v", req)
	var feeLevel bwi.FeeLevel
	switch req.FeeLevel {
	case 0:
		feeLevel = bwi.PRIOIRTY
	case 1:
		feeLevel = bwi.NORMAL
	case 2:
		feeLevel = bwi.ECONOMIC
	case 3:
		feeLevel = bwi.FEE_BUMP
	default:
		return pb.WithdrawResp{}, errors.New("无效的参数")
	}

	txid, err := s.multiWallet.Withdraw(req.Code, req.PropertyId, req.Address, req.Amount, feeLevel)
	if err != nil {
		log.Errorf("server-Withdraw-Error: %s", err.Error())
		return pb.WithdrawResp{}, err
	}
	return pb.WithdrawResp{Txid: txid}, nil
}

// 获取Omni交易信息
func (s *Service) GetOmniTransaction(ctx context.Context, req pb.GetOmniTransactionReq) (pb.GetOmniTransactionResp, error) {
	log.Infof("Visiting GetOmniTransaction, Request Params is %+v", req)
	wallet, _ := s.multiWallet.WalletForCurrencyCode(req.Code)
	bitcoinWallet, ok := wallet.(*bitcoin.BitcoinWallet)
	if !ok {
		return pb.GetOmniTransactionResp{}, errors.New("获取Omni交易失败")
	}

	omniTransaction, err := bitcoinWallet.GetOmniTransaction(req.Txid)
	if err != nil {
		log.Errorf("server-GetOmniTransaction-Error: %s", err.Error())
		return pb.GetOmniTransactionResp{}, err
	}

	// 后续有需要用到的参数，可以进行添加
	transaction := pb.GetOmniTransactionResp{
		Amount: omniTransaction.Amount,
		Fee: omniTransaction.Fee,
		Txid: omniTransaction.TXID,
		BlockTime: omniTransaction.BlockTime,
	}

	return transaction, nil
}