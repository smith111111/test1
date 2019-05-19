package controller

import (
	"galaxyotc/common/data"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/net"
)

// 拉取推送消息
func History(c *gin.Context) {
	to := c.Query("to")
	if to == "" {
		net.SendErrJSON("参数无效", c)
		return
	}

	msgId, err := strconv.Atoi(c.Query("msgid"))
	if err != nil {
		log.Errorf("Im-History-Error: %s", err.Error())
		net.SendErrJSON("参数无效", c)
		return
	}

	userI, _ := c.Get("user")
	user := userI.(model.User)

	var db = model.DB.Model(&model.ImMsg{}).Where("msgIdServer > ? AND (fromAccount = ? AND `to` = ?) OR (fromAccount = ? AND `to` = ?)", msgId, user.ID, to, to, user.ID)

	pushMsgs := []*model.ImMsg{}
	if err := db.Order("msgIdServer desc").Find(&pushMsgs).Error; err != nil {
		log.Errorf("Im-History-Error: %s", err.Error())
		net.SendErrJSON("内部错误", c)
		return
	}

	rsp := []*data.ImHistoryMsgResp{}
	for _, pushMsg := range pushMsgs {
		item := &data.ImHistoryMsgResp{
			MsgId: pushMsg.MsgIdServer,
			MsgType: pushMsg.MsgType,
			Body: pushMsg.Body,
			Attach: pushMsg.Attach,
			Ext: pushMsg.Ext,
			MsgTimestamp: pushMsg.MsgTimestamp,
		}
		rsp = append(rsp, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"list": rsp,
		},
	})
}
