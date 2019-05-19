package service

import (
	"context"
	pb "galaxyotc/common/proto/wallet/eos_wallet"
	"galaxyotc/common/log"
	"github.com/nats-io/go-nats"
	"strconv"
)

type Service struct {
	wallet *EosWallet
}

// 创建新服务
func NewService() *Service {
	service := &Service{}
	service.wallet = &EosWallet{}
	service.wallet.AddTransactionListener(EosCallback)
	return service
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewEosWalletServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName+"_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	s.wallet.Start()

	return nil
}

// 服务关闭
func (s *Service) Close() {
	//s.wallet.Close()
}

// 获取充值地址
func (s *Service) Deposit(ctx context.Context, req pb.DepositReq) (pb.DepositResp, error) {
	log.Infof("Visiting Deposit, Request Params is %+v", req)

	/*
	var purpose wi.KeyPurpose
	switch req.Purpose {
	case 0:
		purpose = wi.EXTERNAL
	case 1:
		purpose = wi.INTERNAL
	default:
		return pb.DepositResp{}, errors.New("无效的参数")
	}

	address, err := s.wallet.Deposit(purpose)
	if err != nil {
		log.Errorf("server-Deposit-Error: %s", err.Error())
		return pb.DepositResp{}, err
	}*/

	address := ""

	return pb.DepositResp{Address: address}, nil
}

// 提现申请：发送ETH给接收者
func (s *Service) EosWithdraw(ctx context.Context, req pb.EosWithdrawReq) (pb.EosWithdrawResp, error) {
	log.Infof("Visiting EtherWithdraw, Request Params is %+v", req)

	amount, err := strconv.Atoi(req.Value)
	if err != nil {
		return pb.EosWithdrawResp{}, err
	}

	txid, err := s.wallet.EosWithdraw(req.To, int64(amount), "")
	if err != nil {
		log.Errorf("server-EtherWithdraw-Error: %s", err.Error())
		return pb.EosWithdrawResp{}, err
	}

	return pb.EosWithdrawResp{Txid: txid}, nil
}
