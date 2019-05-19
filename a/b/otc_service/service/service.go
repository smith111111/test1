package service

import (
	"context"
	pb "galaxyotc/common/proto/backend/otc"
	"galaxyotc/common/log"
	"github.com/nats-io/go-nats"
)

const (
	OrderCancel = 1
	OrderRelease = 2
	OrderCompleted = 3
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
	p2p := pb.NewOTCServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Close() error {
	return s.Close()
}

// 订单超时回调
func (s *Service) OrderTimeoutCallback(ctx context.Context, req pb.OrderTimeoutCallbackReq) (pb.BoolResp, error) {
	log.Infof("Visiting OrderTimeoutCallback, Request Params is %+v", req)
	var (
		ok bool
		err error
	)
	switch req.CallbackType {
	case OrderCancel:
		ok, err = s.db.OrderTimeoutCancelCallback(req.Sn, req.Scene)
	case OrderRelease:
		ok, err = s.db.OrderTimeoutReleaseCallback(req.Sn)
	case OrderCompleted:
		ok, err = s.db.OrderTimeoutCompletedCallback(req.Sn)
	}

	if err != nil {
		log.Errorf("server-OrderTimeoutCallback-Error: %s", err.Error())
		return pb.BoolResp{}, err
	}

	return pb.BoolResp{ok}, nil
}