package push

import (
	_ "github.com/go-sql-driver/mysql"
	"galaxyotc/push_server/model"
	"testing"
)

func TestUMengPush(t *testing.T) {
	req := &model.PushInfo{}
	req.Title = "test1"
	req.Text = "text1"
	req.DisplayType = "notification"
	req.Custom = make(map[string]interface{})
	req.Custom["k1"] = "v1"
	req.Custom["k2"] = "v2"
	req.AppKeySecrets = make(map[string]*model.DeviceTokenAndPackageName)
	req.AppKeySecrets["57c66c66e0f55a02322002e51&mnxcrdldae9anvh7fxng7p4iclw9cbdg"] = &model.DeviceTokenAndPackageName{
		DeviceTokens: "AkrMerwbrJ6kMiCOTo4dLJKsQe5eDLEKiTwsoKl67qLF",
		UserIds:      []uint{10001},
		MsgIds:       []int64{1},
		PackageName:  "",
	}

	ok := GetUmengService().Push(req, model.EPUSH_MSG_TYPE_LISTCAST, model.EDEVICE_TYPE_ANDROID, func(changeStatus, pushId, errorCode string) {
		t.Log("10002", changeStatus, pushId, errorCode)
	})

	t.Log(ok)
}
