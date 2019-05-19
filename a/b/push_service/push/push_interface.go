package push

import (
	"galaxyotc/common/model"
)

type PushDoneCallBack func(changeStatus, pushId, errorCode string)

type IPush interface {
	Push(req *model.PushInfo, pushMsgType model.EPUSH_MSG_TYPE, deviceType model.EDEVICE_TYPE, cb PushDoneCallBack) bool
}
