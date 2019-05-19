package service

import (
	pb "galaxyotc/common/proto/push"
	"galaxyotc/common/data"
	"context"
	"galaxyotc/common/log"
	"errors"
	"galaxyotc/common/model"
)


// 验证请求参数
func verifySendMsgParam(req *pb.SendMsgReq) bool {
	if len(req.Receivers) == 0 {
		return false
	}

	loginStatus := req.LoginStatus

	if loginStatus != model.ELOGIN_STATUS_NOLOGIN &&
		loginStatus != model.ELOGIN_STATUS_LOGIN &&
		loginStatus != model.ELOGIN_STATUS_ALL {
		loginStatus = model.ELOGIN_STATUS_LOGIN
	}

	if req.DisplayType != model.DISPLAY_TYPE_NOTIFICATION &&
		req.DisplayType != model.DISPLAY_TYPE_MESSAGE {
		req.DisplayType = model.DISPLAY_TYPE_NOTIFICATION
	}

	return true
}

type PushP2PService struct {
	PS *PushService
}

// 创建新服务
func NewPushP2PService() *PushP2PService {
	service := &PushP2PService{
		PS: newPushService(),
	}
	return service
}

// 用户登出
func (s *PushP2PService) Logout(ctx context.Context, req pb.LogoutReq) (pb.LogoutResp, error) {
	log.Infof("Visiting Register, Request Params is %+v", req)

	ok := s.PS.TS.Logout(req.UserId, req.AppId)
	if !ok {
		return pb.LogoutResp{int32(data.ErrorCode.ERROR), "fail"}, errors.New("用户登出失败")
	}

	return pb.LogoutResp{int32(data.ErrorCode.ERROR), "success"}, nil
}

// 单播, 列播
func (s *PushP2PService) SendMsg(ctx context.Context, req pb.SendMsgReq) (pb.SendMsgResp, error) {
	log.Infof("Visiting Register, Request Params is %+v", req)

	ok := verifySendMsgParam(&req)
	if !ok {
		return pb.SendMsgResp{int32(data.ErrorCode.ERROR), "fail"}, errors.New("参数无效")
	}

	ok, msg := s.PS.Push(&req)
	if !ok {
		return pb.SendMsgResp{int32(data.ErrorCode.ERROR), msg}, errors.New("发送消息失败")
	}

	return pb.SendMsgResp{int32(data.ErrorCode.ERROR), "success"}, nil
}