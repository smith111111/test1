package controller

import (
	"net/http"
	"encoding/json"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/data"
	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"

	"galaxyotc/gc_services/im_service/client"
	"strconv"
)

// 消息抄送
func ReceiveMsg(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	body, err := client.ImClient.GetEventNotification(c.Request)
	if err != nil {
		log.Errorf("Im-ReceiveMsg-Error: %s", err.Error())
		SendErrJSON("消息抄送失败", c)
		return
	}

	resp := &data.MsgCopyInfoResp{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		log.Errorf("Im-ReceiveMsg-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	msgId, _ := strconv.ParseInt(resp.MsgIdServer, 10, 64)
	msgTimestramp, _ := strconv.ParseInt(resp.MsgTimestamp, 10, 64)
	eventType, _ := strconv.Atoi(resp.EventType)
	resendFlag, _ := strconv.Atoi(resp.ResendFlag)

	pm := &model.ImMsg{
		MsgIdServer: msgId,
		MsgIdClient: resp.MsgIdClient,
		EventType: int8(eventType),
		ConvType: resp.ConvType,
		To: resp.To,
		FromAccount: resp.FromAccount,
		FromClientType: resp.FromClientType,
		FromDeviceId: resp.FromDeviceId,
		FromNick: resp.FromNick,
		MsgTimestamp: msgTimestramp,
		MsgType: resp.MsgType,
		Body: resp.Body,
		Attach: resp.Attach,
		ResendFlag: int8(resendFlag),
		CustomApnsText: resp.CustomApnsText,
		Ext: resp.Ext,
	}

	if err := model.DB.Create(&pm).Error; err != nil {
		log.Errorf("Im-ReceiveMsg-Error: %s", err.Error())
		SendErrJSON("服务器出错啦！", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
