package controller

import (
	"net/http"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/data"
	"galaxyotc/common/model"

	"galaxyotc/gc_services/push_service/auth"
	"galaxyotc/gc_services/push_service/service"

	"github.com/gin-gonic/gin"
	"galaxyotc/gc_services/push_service/api"
)

// 上报设备相关信息
func DeviceInfo(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	appKey := c.Request.Header.Get("AppKey")
	_, ok := auth.GetAuthAppKeyService().GetSecretKey(appKey)
	if !ok {
		SendErrJSON("AppKey有误", c)
		return
	}

	req := &data.PushDeviceInfoReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		log.Errorf("Push-DeviceInfo-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if !verifyDeviceInfo(req) {
		SendErrJSON("参数无效", c)
		return
	}

	userI, _ := c.Get("user")
	user := userI.(model.User)

	platform := model.EDEVICE_TYPE_ANDROID
	if appKey == data.APPKEY_OTC_IOS {
		platform = model.EDEVICE_TYPE_IOS
	} else if appKey == data.APPKEY_OTC_H5 {
		platform = model.EDEVICE_TYPE_H5
	}

	api.PushApi.BroadcastDeviceInfo(req.AppId, req.DeviceToken, req.PushType, user.ID, platform)

	ok, msg := service.PService.P2P.PS.TS.Save(req, uint64(user.ID), platform)
	if !ok {
		SendErrJSON(msg, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
	})
}

func verifyDeviceInfo(req *data.PushDeviceInfoReq) bool {
	if req.AppId != data.APPID_OTC {
		return false
	}
	return true
}
