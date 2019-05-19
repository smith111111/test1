package controller

import (
	"fmt"
	"time"
	"net/http"

	"galaxyotc/common/log"
	"galaxyotc/common/net"
	"galaxyotc/common/data"
	"galaxyotc/common/model"
	pb "galaxyotc/common/proto/push"

	"galaxyotc/gc_services/push_service/service"

	"github.com/gin-gonic/gin"
)

// 用户发推送消息
func SendMessage(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	req := &data.PushSendMsgReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		log.Errorf("Push-SendMessage-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	userI, _ := c.Get("user")
	user := userI.(model.User)

	newReq := &pb.SendMsgReq{
		DisplayType: "notification",
		Receivers: req.Receivers,
		AppId: data.APPID_OTC,
		Title: fmt.Sprintf("%s给您发了消息", user.Name),
		Text: req.Text,
		Custom: []byte("{}"),
		ExpireTime: int32(time.Now().AddDate(0, -3, 0).Unix()),
		LoginStatus: 1,
	}

	ok, msg := service.PService.P2P.PS.Push(newReq)
	if !ok {
		log.Error(msg)
		SendErrJSON(msg, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg": "success",
	})
}