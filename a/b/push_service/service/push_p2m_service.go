package service

import (
	pb "galaxyotc/common/proto/push"
	"galaxyotc/common/data"
	"context"
	"galaxyotc/common/log"
)

type PushP2MService struct {
	ts *TokenService
}

// 创建新服务
func NewPushP2MService() *PushP2MService {
	service := &PushP2MService{
		ts: newTokenService(),
	}
	return service
}

// 上报设备同步
func (s *PushP2MService) SyncDeviceInfo(ctx context.Context, req pb.DeviceInfoReq) (rsp pb.DeviceInfoResp, err error) {
	log.Infof("Visiting Register, Request Params is %+v", req)

	s.ts.SaveCache(int8(req.AppId), req.DeviceToken, req.PushType, req.Uid, req.Platform)

	rsp.ErrNo = int32(data.ErrorCode.SUCCESS)
	rsp.Msg = "success"

	return rsp, nil
}