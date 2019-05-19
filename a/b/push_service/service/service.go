package service

import (
	"context"
	pb "galaxyotc/common/proto/push"

	"github.com/nats-io/go-nats"
)

var PService *Service

type Service struct {
	P2P *PushP2PService
	P2M *PushP2MService
}

// 创建新服务
func NewService() {
	service := &Service{
		P2P: NewPushP2PService(),
		P2M: NewPushP2MService(),
	}
	PService = service
}

func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewPushP2PHandler(context.Background(), nc, s.P2P)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}

	p2m := pb.NewPushP2MHandler(context.Background(), nc, s.P2M)
	_, err = nc.Subscribe(p2m.Subject(), p2m.Handler)
	if err != nil {
		return err
	}
	return nil
}