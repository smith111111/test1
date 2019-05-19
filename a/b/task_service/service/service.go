package service

import (
	"context"

	pb "galaxyotc/common/proto/task"
	"galaxyotc/common/log"

	"github.com/RussellLuo/timingwheel"
	"time"
	"github.com/nats-io/go-nats"
	"galaxyotc/gc_services/task_service/api"
)

type Service struct {
	timingWheel *timingwheel.TimingWheel
}

// 创建新服务
func NewService() *Service {
	return &Service{
		timingWheel: timingwheel.NewTimingWheel(time.Second * 5, 10000),
	}
}

// 服务启动
func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewTaskServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	s.timingWheel.Start()
	return nil
}

// 服务关闭
func (s *Service) Close() error {
	s.timingWheel.Stop()
	return s.Close()
}

// 订单超时取消回调
func (s *Service) OrderTimeout(ctx context.Context, req pb.OrderTimeoutReq) (pb.BoolResp, error) {
	log.Infof("Visiting OrderTimeout, Request Params is %+v", req)
	s.timingWheel.AfterFunc(time.Duration(req.Duration), func() {
		if _, err := api.OTCApi.OrderTimeoutCallback(req.CallbackType, req.OrderSn, req.Scene); err != nil {
			log.Errorf("service-OrderTimeout-error: %s", err.Error())
		}
	})

	return pb.BoolResp{true}, nil
}

// 账户转账交易
func (s *Service) AccountTransfer(ctx context.Context, req pb.AccountTransferReq) (pb.BoolResp, error) {
	log.Infof("Visiting AccountTransfer, Request Params is %+v", req)
	if _, err := api.AccountApi.AccountTransfer(req.Amount, req.Sn, req.ReceiverAddress, req.SenderAddress, req.TokenAddress); err != nil {
		log.Errorf("service-AccountTransfer-error: %s", err.Error())
		return pb.BoolResp{false}, err
	}

	return pb.BoolResp{true}, nil
}