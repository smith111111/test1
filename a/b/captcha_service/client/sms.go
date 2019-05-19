package client

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"galaxyotc/common/log"
)

// 发送短信
func SendSMSMsg(reqURL string) (err error) {
	if reqURL == "" {
		return errors.New("请求地址不能为空")
	}

	var res *http.Response
	log.Infof("Request Url is %s", reqURL)
	res, err = http.DefaultClient.Get(reqURL)
	if err != nil {
		log.Errorf("SendSMSMsg Error: %s", err.Error())
		return errors.New("发送短信验证码失败")
	}
	defer res.Body.Close()

	resBody, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		log.Errorf("SendSMSMsg Error: %s", readErr.Error())
		return errors.New("发送短信验证码失败")
	}

	type SMSResult struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}

	var smsResult SMSResult
	if err := json.Unmarshal(resBody, &smsResult); err != nil {
		log.Errorf("SendSMSMsg Error: %s", err.Error())
		return errors.New("发送短信验证码失败")
	}

	if smsResult.Code != 1 {
		log.Errorf("smsResult Error: %+v", smsResult)
		return errors.New("发送短信验证码失败")
	}

	return nil

}
