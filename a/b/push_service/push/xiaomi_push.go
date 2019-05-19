package push

import (
	"crypto/tls"
	"encoding/json"
	"galaxyotc/common/model"
	"strconv"

	"strings"

	"github.com/astaxie/beego/httplib"
	"galaxyotc/common/log"
)

//推送返回结果
type XiaoMiPushRsp struct {
	Result      string            `json:"result,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	TraceId     string            `json:"trace_id,omitempty"`
	Code        int32             `json:"code,omitempty"`
	Data        XiaoMiPushRspData `json:"data,omitempty"`
	Description string            `json:"description,omitempty"`
	Info        string            `json:"info,omitempty"`
}

type XiaoMiPushRspData struct {
	Id string `json:"id,omitempty"`
}

//测试返回的数据
//{"result":"error","reason":"No valid targets!","trace_id":"Xcm07b77488763663927bm","code":20301,"description":"发送消息失败"}
//{"result":"ok","trace_id":"Xlm12b34488763708227kX","code":0,"data":{"id":"alm12b34488763708232Vy"},"description":"成功","info":"Received push messages for 1 ALIAS"}

var xiaomi *XiaoMiPushService

func GetXiaoMiService() *XiaoMiPushService {
	if xiaomi == nil {
		xiaomi = &XiaoMiPushService{}
	}
	return xiaomi
}

type XiaoMiPushService struct{}

func (p *XiaoMiPushService) Push(req *model.PushInfo, pushMsgType model.EPUSH_MSG_TYPE, deviceType model.EDEVICE_TYPE, cb PushDoneCallBack) bool {
	log.Debugf("PushInfo:%v", req)

	for k, v := range req.AppKeySecrets {
		key := strings.Split(k, "&")
		secret := key[1]

		b := httplib.Post(model.XIAOMI_PUSH_URL)
		b.Header("Authorization", "key="+secret)
		b.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

		vs := p.makeBody(req, v)
		for k, v := range vs {
			tmp, ok := v.(string)
			if ok {
				b.Param(k, tmp)
			} else {
				s, err := json.Marshal(v)
				if err == nil {
					b.Param(k, string(s))
				}
			}
		}

		json.Marshal(vs)
		log.Debug(vs)

		rsp := &XiaoMiPushRsp{}
		if err := b.ToJSON(rsp); err != nil {
			log.Errorf("err: %s", err.Error())
			return false
		}

		log.Debug(rsp)

		if rsp.Result != "ok" {
			return false
		}

		if cb != nil {
			cb(rsp.Result, rsp.Data.Id, strconv.Itoa(int(rsp.Code)))
		}

		log.Infof("push success~")
	}

	return true
}

func (p *XiaoMiPushService) makeBody(req *model.PushInfo, v *model.DeviceTokenAndPackageName) map[string]interface{} {
	vs := make(map[string]interface{})

	vs["restricted_package_name"] = v.PackageName
	vs["registration_id"] = v.DeviceTokens

	if req.DisplayType == "notification" {
		vs["pass_through"] = "0"
	} else {
		vs["pass_through"] = "1"
	}

	vs["title"] = req.Title
	vs["description"] = req.Text
	vs["notify_type"] = "-1"
	vs["time_to_live"] = "1000"
	vs["notify_id"] = "0"

	vs["extra"] = req.Custom
	vs["payload"] = req.Custom

	return vs
}
