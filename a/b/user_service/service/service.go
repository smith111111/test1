package service

import (
	"context"
	pb "galaxyotc/common/proto/backend/user"
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
	p2p := pb.NewUserServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Close() error {
	return s.Close()
}

// 获取用户信息
func (s *Service) GetUser(ctx context.Context, req pb.GetUserReq) (pb.UserInfo, error) {
	log.Infof("Visiting GetUser, Request Params is %+v", req)
	resp, err := s.db.GetUser(req.UserId)
	if err != nil {
		log.Errorf("server-GetUser-Error: %s", err.Error())
		return pb.UserInfo{}, err
	}
	return *resp, nil
}

// 根据内部地址获取用户信息
func (s *Service) GetUserByInternalAddress(ctx context.Context, req pb.GetUserByInternalAddressReq) (pb.UserInfo, error) {
	log.Infof("Visiting GetUserByInternalAddress, Request Params is %+v", req)
	resp, err := s.db.GetUserByInternalAddress(req.InternalAddress)
	if err != nil {
		log.Errorf("server-GetUserByInternalAddress-Error: %s", err.Error())
		return pb.UserInfo{}, err
	}
	return *resp, nil
}

// 检查用户是否是否存在
func (s *Service) IsExist(ctx context.Context, req pb.IsExistReq) (pb.IsExistResp, error) {
	log.Infof("Visiting IsExist, Request Params is %+v", req)
	exist, err := s.db.IsExist(req.Input)
	if err != nil {
		log.Errorf("server-IsExist-Error: %s", err.Error())
		return pb.IsExistResp{}, err
	}
	return pb.IsExistResp{exist}, nil
}
