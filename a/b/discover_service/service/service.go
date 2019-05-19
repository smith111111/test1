package service

import (
	"context"
	pb "galaxyotc/common/proto/backend/otc"
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
	//s
	p2p := pb.NewOTCServiceHandler(context.Background(), nc, nil)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Close() error {
	return s.Close()
}
