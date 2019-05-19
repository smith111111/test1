package router

import (
	"galaxyotc/gc_services/push_service/controller"
	"github.com/gin-gonic/gin"
	"galaxyotc/common/middleware"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("push_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 上报设备相关信息
		api.POST("/device/info", middleware.SigninRequired, controller.DeviceInfo)
		// 拉取推送消息
		api.GET("/message/list", middleware.SigninRequired, controller.MessageList)
		// 用户发推送消息
		api.GET("/send/message", middleware.SigninRequired, controller.SendMessage)
	}
}
