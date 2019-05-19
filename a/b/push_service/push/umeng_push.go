package push

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"galaxyotc/common/model"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"galaxyotc/common/log"
)

//推送返回结果
type UmengPushRsp struct {
	Ret  string            `json:"ret"`
	Data *UmengPushRspData `json:"data"`
}

type UmengPushRspData struct {
	MsgId     string `json:"msg_id"` //消息id
	Errorcode string `json:"error_code"`
}

var umeng *UmengPushService

func GetUmengService() *UmengPushService {
	if umeng == nil {
		umeng = &UmengPushService{}
	}
	return umeng
}

type UmengPushService struct{}

func (p *UmengPushService) Push(req *model.PushInfo, pushMsgType model.EPUSH_MSG_TYPE, deviceType model.EDEVICE_TYPE, cb PushDoneCallBack) bool {
	for k, v := range req.AppKeySecrets {
		//取出appkey和对应token
		key := strings.Split(k, "&")
		appkey := key[0]
		secret := key[1]

		var body map[string]interface{}

		if deviceType == model.EDEVICE_TYPE_ANDROID {
			body = p.makeAndroidBody(appkey, v, req, pushMsgType)
		} else {
			body = p.makeIOSBody(appkey, v, req, pushMsgType)
		}

		//将消息发布
		pushItem := p.pushMessage(secret, body)
		if pushItem == nil {
			log.Errorf("友盟推送消息连接失败")
			return false
		}

		log.Debug(pushItem)
		log.Debug(pushItem.Data)

		if pushItem.Ret != "SUCCESS" {
			log.Warnf("推送失败，错误码:%s", pushItem.Data.Errorcode)
			return false
		}

		if cb != nil {
			cb(pushItem.Ret, pushItem.Data.MsgId, pushItem.Data.Errorcode)
		}

		log.Infof("push success~")
	}

	return true
}

func (p *UmengPushService) pushTypeSwap(pushMsgType model.EPUSH_MSG_TYPE) string {
	var ret string

	if pushMsgType == model.EPUSH_MSG_TYPE_LISTCAST {
		ret = model.EPUSH_MSG_TYPE_LISTCAST_STR
	}

	return ret
}

//推送消息
func (p *UmengPushService) pushMessage(appSecret string, body map[string]interface{}) *UmengPushRsp {
	//将body序列化
	data, err := json.Marshal(body)
	if err != nil {
		log.Errorf("json.Marshal fail ,err: %v", err.Error())
		return nil
	}

	log.Debug(string(data))

	//得到签名sign
	sign := p.validUrlSign("POST", model.UMENG_PUSH_URL, string(data), appSecret)

	url := fmt.Sprintf("%s?sign=%s", model.UMENG_PUSH_URL, sign)

	//发请求
	b := httplib.Post(url)
	b.Body(data)

	rsp := &UmengPushRsp{}
	if err := b.ToJSON(rsp); err != nil {
		log.Errorf("post to umeng and make data to json fail ,err : %s", err.Error())
		return nil
	}

	fmt.Println(rsp.Data)

	return rsp
}

func (p *UmengPushService) makeAndroidBody(appKey string, v *model.DeviceTokenAndPackageName, req *model.PushInfo, pushMsgType model.EPUSH_MSG_TYPE) map[string]interface{} {
	log.Debug("makeAndroidBody...")

	//构建post的body
	body := make(map[string]interface{})

	body["appkey"] = appKey
	body["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())                            //当前时间戳
	body["production_mode"] = beego.AppConfig.DefaultString("production_mode", "false") // 正式（true）/测试（false）模式
	body["description"] = req.Title                                                     //消息描述（不显示在通知栏），目前默认是消息的标题

	newPushType := p.pushTypeSwap(pushMsgType)
	body["type"] = newPushType
	if newPushType == "listcast" {
		body["device_tokens"] = v.DeviceTokens
	}

	payload := make(map[string]interface{})   //消息内容
	payload["display_type"] = req.DisplayType //消息类型：notification-通知，message-消息

	msgbody := make(map[string]interface{}) //消息体
	msgbody["ticker"] = req.Title           //消息提示文字
	msgbody["title"] = req.Title            //消息标题
	msgbody["text"] = req.Text              //消息描述
	msgbody["after_open"] = "go_custom"     //点击的行为, 自定义，由客户端自行定义反应

	//特殊处理
	/*if req.AppId == 1 {
		if req.Custom["router"] != nil {
			msgbody["custom"] = req.Custom["router"].(string)
		}

	} else {*/
	msgbody["custom"] = req.Custom //自定义的内容
	//}

	payload["body"] = msgbody

	body["payload"] = payload

	if req.ExpireTime != "" {
		policy := make(map[string]interface{}) //发送策略
		policy["expire_time"] = req.ExpireTime
		body["policy"] = policy
	}

	return body
}

func (p *UmengPushService) makeIOSBody(appKey string, v *model.DeviceTokenAndPackageName, req *model.PushInfo, pushType model.EPUSH_MSG_TYPE) map[string]interface{} {
	log.Debug("makeIOSBody...")

	//构建post的body
	body := make(map[string]interface{})
	body["appkey"] = appKey
	body["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())                            //当前时间戳
	body["production_mode"] = beego.AppConfig.DefaultString("production_mode", "false") // 正式（true）/测试（false）模式
	body["description"] = req.Title                                                     //消息描述（不显示在通知栏），目前默认是消息的标题
	// body["thirdparty_id"] = time.Now().Format("2006-01-02 15:04:05")                    //消息的唯一标识

	newPushType := p.pushTypeSwap(pushType)
	body["type"] = newPushType //列播，小于500个token，可以为1个
	if newPushType == "listcast" {
		body["device_tokens"] = v.DeviceTokens
	}

	payload := make(map[string]interface{}) //消息内容
	aps := make(map[string]interface{})

	if req.Text == "" {
		req.Text = "消息"
	}

	/*if req.AppId == 1 || req.AppId == 6 {
		alert := make(map[string]interface{})
		alert["title"] = req.Title
		alert["body"] = req.Text
		aps["alert"] = alert
		aps["sound"] = "default"

		if req.Custom["router"] != nil {
			aps["router"] = req.Custom["router"].(string)
		}

		payload["aps"] = aps
	} else {*/
	//自定义的行为，由客户端处理
	aps["alert"] = req.Text //消息内容  （iOS标题固定为指尖遥控iOS）
	aps["sound"] = "default"
	payload["aps"] = aps

	for key, value := range req.Custom {
		payload[key] = value
	}
	//}

	body["payload"] = payload

	if req.ExpireTime != "" {
		policy := make(map[string]interface{}) //发送策略
		policy["expire_time"] = req.ExpireTime
		body["policy"] = policy
	}

	return body
}

//url签名
func (p *UmengPushService) validUrlSign(method, url, post_body, app_secret string) string {
	data := fmt.Sprintf("%s%s%s%s", method, url, post_body, app_secret)
	//计算md5
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(data))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
