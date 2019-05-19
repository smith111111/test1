package push

import (
	"galaxyotc/common/model"
	"galaxyotc/common/log"
	"strconv"
	"galaxyotc/gc_services/push_service/api"
)

//推送返回结果
type GalaxyPushRsp struct {
	Result string `json:"result,omitempty"`
}

var galaxy *GalaxyPushService

func GetGalaxyService() *GalaxyPushService {
	if galaxy == nil {
		galaxy = &GalaxyPushService{}
	}
	return galaxy
}

type GalaxyPushService struct {
}

func (p *GalaxyPushService) Push(req *model.PushInfo, pushType model.EPUSH_MSG_TYPE, deviceType model.EDEVICE_TYPE, cb PushDoneCallBack) bool {
	for _, v := range req.AppKeySecrets {
		for i, userId := range v.UserIds {
			msgId := v.MsgIds[i]

			err := api.FrontApi.RecvPush(uint64(userId), msgId, req.Title, req.Text, req.Custom)
			if err != nil {
				log.Error("推送失败，错误码:%s", err)
			}

			var errStr string

			if err != nil {
				errStr = err.Error()
			}

			if cb != nil {
				cb("", strconv.Itoa(int(msgId)), errStr)
			}
		}

		log.Infof("push success~")
	}

	return true
}
