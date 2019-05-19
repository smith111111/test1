package push

import (
	"galaxyotc/common/model"
)

var m map[model.EPUSH_SYS_TYPE]IPush

func init() {
	m = make(map[model.EPUSH_SYS_TYPE]IPush)
	m[model.EDEVICE_SYS_TYPE_UMENG] = GetUmengService()
	m[model.EDEVICE_SYS_TYPE_JIGUANG] = GetJiGuangService()
	m[model.EDEVICE_SYS_TYPE_XIAOMI] = GetXiaoMiService()
	m[model.EDEVICE_SYS_TYPE_GALAXY] = GetGalaxyService()
}

func GetPushService(pushSysType model.EPUSH_SYS_TYPE) IPush {
	p, ok := m[pushSysType]
	if !ok {
		return nil
	}

	return p
}
