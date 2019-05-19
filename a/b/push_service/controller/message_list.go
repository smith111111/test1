package controller

import (
	"galaxyotc/common/data"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"galaxyotc/common/model"
	"galaxyotc/common/net"
	"galaxyotc/common/log"
)

// 拉取推送消息
func MessageList(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	msgTypes := strings.Split(c.Query("msg_types"), ",")
	if len(msgTypes) == 0 {
		SendErrJSON("参数无效", c)
		return
	}

	msgId, err := strconv.Atoi(c.Query("msgid"))
	if err != nil {
		log.Errorf("Push-MessageList-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	userI, _ := c.Get("user")
	user := userI.(model.User)

	db := model.DB.Model(&model.PushMsg{}).Where("msgid > ? AND userid = ? AND msg_type in (?)", msgId, user.ID, msgTypes).Order("msgid desc").Offset(0).Limit(100)

	pushMsgs := []*model.PushMsg{}
	if err := db.Find(&pushMsgs).Error; err != nil {
		log.Errorf("Push-MsgList-Error: %s", err.Error())
		SendErrJSON("内部错误", c)
		return
	}

	resp := []*data.PushPushMsgResp{}
	for _, pushMsg := range pushMsgs {
		item := &data.PushPushMsgResp{
			MsgId: pushMsg.MsgId,
			Title: pushMsg.Title,
			Text: pushMsg.Text,
			Custom: pushMsg.Custom,
		}

		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"list": resp,
		},
	})
}
