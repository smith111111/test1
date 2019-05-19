package push

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"galaxyotc/common/model"
	"strings"
	"github.com/astaxie/beego/httplib"
	"galaxyotc/common/log"
	"github.com/spf13/viper"
)

type JiGuangPushRsp struct {
	SendNo string `json:"sendno,omitempty"`
	MsgId  string `json:"msg_id,omitempty"`
}

type JiGuangPushService struct {
}

var jiguang *JiGuangPushService

func GetJiGuangService() *JiGuangPushService {
	if jiguang == nil {
		jiguang = &JiGuangPushService{}
	}
	return jiguang
}

func (p *JiGuangPushService) Push(req *model.PushInfo, pushType model.EPUSH_MSG_TYPE, deviceType model.EDEVICE_TYPE, cb PushDoneCallBack) bool {
	for k, v := range req.AppKeySecrets {
		key := strings.Split(k, "&")
		auth := key[0] + ":" + key[1]

		auth = base64.StdEncoding.EncodeToString([]byte(auth))

		b := httplib.Post(model.JIGUANG_PUSH_URL)
		b.Header("Content-Type", "application/json")
		b.Header("Authorization", "Basic "+auth)
		b.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

		body := p.makeBody(req, v)
		if body == nil {
			return false
		}

		data, err := json.Marshal(body)
		if err != nil {
			log.Errorf("json.Marshal fail ,err: %v", err.Error())
			return false
		}

		b.Body(data)
		buf, err := b.Bytes()
		if err != nil {
			log.Errorf("json.Marshal fail ,err: %v", err.Error())
			return false
		}

		rsp := &JiGuangPushRsp{}
		err = json.Unmarshal(buf, rsp)
		if err != nil {
			log.Errorf("json.Marshal fail ,err: %v", err.Error())
			return false
		}

		if rsp.SendNo != "0" {
			return false
		}

		if cb != nil {
			cb(rsp.SendNo, rsp.MsgId, "0")
		}

		log.Infof("push success~")
	}

	return true
}

func (p *JiGuangPushService) makeBody(req *model.PushInfo, v *model.DeviceTokenAndPackageName) map[string]interface{} {
	body := make(map[string]interface{})
	body["platform"] = "all"

	audience := make(map[string]interface{})
	audience["registration_id"] = strings.Split(v.DeviceTokens, ",")
	body["audience"] = audience

	extras, err := json.Marshal(req.Custom)
	if err != nil {
		return nil
	}

	if req.DisplayType == "notification" {
		android := make(map[string]interface{})
		android["alert"] = req.Text
		android["title"] = req.Title
		android["builder_id"] = 1
		android["extras"] = req.Custom

		ios := make(map[string]interface{})
		ios["alert"] = req.Text
		ios["sound"] = "default"
		ios["badge"] = "+1"
		ios["extras"] = req.Custom

		notification := make(map[string]interface{})
		notification["android"] = android
		notification["ios"] = ios
		body["notification"] = notification

	} else {
		message := make(map[string]interface{})
		message["msg_content"] = string(extras)
		message["content_type"] = "text"
		message["title"] = req.Title
		//message["extras"] = req.Custom

		body["message"] = message
	}

	options := make(map[string]interface{})
	options["time_to_live"] = 60

	if viper.GetBool("server.dev") {
		options["apns_production"] = false
	} else {
		options["apns_production"] = true
	}

	body["options"] = options

	return body
}
