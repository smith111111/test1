package push

import (
	_ "github.com/go-sql-driver/mysql"
	"galaxyotc/push_server/model"
	"testing"
)

func TestJiGuangPush(t *testing.T) {
	req := &model.PushInfo{}
	req.Title = "test1"
	req.Text = "text1"
	req.Custom = make(map[string]interface{})
	req.Custom["k1"] = "v1"
	req.Custom["k2"] = "v2"
	req.AppKeySecrets = make(map[string]*model.DeviceTokenAndPackageName)
	req.AppKeySecrets["40e831c0548b8d981b787dd6&3cf9c572f75550fcd4dddf44"] = &model.DeviceTokenAndPackageName{
		DeviceTokens: "170976fa8add4a6097b",
		UserIds:      []uint{10001},
		MsgIds:       []int64{1},
		PackageName:  "",
	}

	ok := GetJiGuangService().Push(req, model.EPUSH_MSG_TYPE_LISTCAST, model.EDEVICE_TYPE_ANDROID, func(changeStatus, pushId, errorCode string) {
		t.Log("10001", changeStatus, pushId, errorCode)
	})

	t.Log(ok)
}
