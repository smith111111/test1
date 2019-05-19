package service

import (
	"context"
	pb "galaxyotc/common/proto/im"

	"galaxyotc/common/log"
	"github.com/spf13/viper"
	"github.com/gao88/netease-im"
	"galaxyotc/common/data"
	"strconv"
	"github.com/nats-io/go-nats"
)

type Service struct {
	*netease.ImClient
}

// 创建新服务
func NewService() *Service {
	service := &Service{
		netease.CreateImClient(viper.GetString("im_service.app_key"), viper.GetString("im_service.app_secret"), ""),
	}
	return service
}

func (s *Service) Start(serviceName string, nc *nats.Conn) error {
	p2p := pb.NewIMServiceHandler(context.Background(), nc, s)
	_, err := nc.QueueSubscribe(p2p.Subject(), serviceName + "_p2p", p2p.Handler)
	if err != nil {
		return err
	}
	return nil
}

// 上报设备相关信息
func (s *Service) UserRegister(ctx context.Context, req pb.UserRegisterReq) (pb.UserRegisterResp, error) {
	log.Infof("Visiting Register, Request Params is %+v", req)

	id := strconv.FormatUint(req.Id, 10)

	user := &netease.ImUser{
		ID:        id,
		Name:      req.Name,
		Propertys: string(req.Props),
		IconURL:   req.Icon,
		Sign:      "",
		Email:     req.Email,
		Birthday:  req.Birth,
		Mobile:    req.Mobile,
		Gender:    int(req.Gender),
		Extension: string(req.Ex),
	}

	tk, err := s.ImClient.CreateImUser(user)
	if err != nil {
		log.Errorf("server-Register-Error: %s", err.Error())
		return pb.UserRegisterResp{int32(data.ErrorCode.ERROR), err.Error(), ""}, err
	}

	return pb.UserRegisterResp{int32(data.ErrorCode.SUCCESS), "success", tk.Token}, nil
}

func (s *Service) UserUpdate(ctx context.Context, req pb.UserUpdateReq) (rep pb.UserUpdateResp, err error) {
	return pb.UserUpdateResp{}, nil
}

func (s *Service) SendSysMsg(ctx context.Context, req pb.SendSysMsgReq) (rsp pb.SendSysMsgResp, err error) {
	log.Infof("Visiting SendSysMsg, Request Params is %+v", req)
	from := strconv.Itoa(int(req.From))
	to := strconv.Itoa(int(req.To))

	opt := &netease.ImSendAttachMessageOption{
		Pushcontent: "",
		Payload: "",
		Sound: "",
		Save: 1,
	}

	err = s.ImClient.SendAttachMsg(from, to, string(req.Attach), opt)
	if err != nil {
		log.Errorf("server-SendSysMsg-Error: %s", err.Error())
		return pb.SendSysMsgResp{int32(data.ErrorCode.ERROR), "内部错误"}, err
	}

	return pb.SendSysMsgResp{int32(data.ErrorCode.SUCCESS), "success"}, nil
}