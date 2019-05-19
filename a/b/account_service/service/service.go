package service

import (
	"context"
	pb "galaxyotc/common/proto/backend/account"
	"galaxyotc/common/log"
	"github.com/nats-io/go-nats"
)

type Service struct {
	*db
}

// 创建新服务
func NewService() *Service {
	service := &Service{
		db: newDB(),
	}
	return service
}

func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewAccountServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

// EthereumWallet监听回调
func (s *Service) EthereumCallback(ctx context.Context, req pb.CallbackReq) (pb.BoolResp, error) {
	log.Infof("Visiting EthereumCallback, Request Params is %+v", req)
	s.db.EthereumCallback(req)
	return pb.BoolResp{true}, nil
}

// EosWallet监听回调
func (s *Service) EosCallback(ctx context.Context, req pb.CallbackReq) (pb.BoolResp, error) {
	log.Infof("Visiting EosCallback, Request Params is %+v", req)
	s.db.EosCallback(req)
	return pb.BoolResp{true}, nil
}

// MultiWallet监听回调
func (s *Service) MultiCallback(ctx context.Context, req pb.CallbackReq) (pb.BoolResp, error) {
	log.Infof("Visiting MultiCallback, Request Params is %+v", req)
	s.db.MultiCallback(req)
	return pb.BoolResp{true}, nil
}

// PrivateWallet 失败处理回调
func (s *Service) PrivateErrorCallback(ctx context.Context, req pb.CallbackReq) (pb.BoolResp, error) {
	log.Infof("Visiting PrivateErrorCallback, Request Params is %+v", req)
	s.db.PrivateErrorCallback(req)
	return pb.BoolResp{true}, nil
}

// 账户转账交易
func (s *Service) AccountTransfer(ctx context.Context, req pb.AccountTransferReq) (pb.BoolResp, error) {
	log.Infof("Visiting AccountTransfer, Request Params is %+v", req)
	s.db.AccountTransfer(req)
	return pb.BoolResp{true}, nil
}