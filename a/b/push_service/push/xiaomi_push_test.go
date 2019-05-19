package push

import (
	_ "github.com/go-sql-driver/mysql"
	"galaxyotc/push_server/model"
	"testing"
)

func TestXiaoMiPush(t *testing.T) {
	req := &model.PushInfo{}
	req.Title = "test1"
	req.Text = "text1"
	req.Custom = make(map[string]interface{})
	req.Custom["k1"] = "v1"
	req.Custom["k2"] = "v2"
	req.AppKeySecrets = make(map[string]*model.DeviceTokenAndPackageName)
	req.AppKeySecrets["5181754939984&QLzoCrqaSddvZ3qGgeWgAA=="] = &model.DeviceTokenAndPackageName{
		DeviceTokens: "AxuJt/wMOaIGr8p+zDt1R4BddTtHV1mSMwXthgPEbac=",
		UserIds:      []uint{10001},
		MsgIds:       []int64{1},
		PackageName:  "",
	}

	ok := GetXiaoMiService().Push(req, model.EPUSH_MSG_TYPE_LISTCAST, model.EDEVICE_TYPE_ANDROID, func(changeStatus, pushId, errorCode string) {
		t.Log("10001", changeStatus, pushId, errorCode)
	})

	t.Log(ok)
}
