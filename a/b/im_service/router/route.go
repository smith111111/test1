package router

import (
	"galaxyotc/gc_services/im_service/controller"
	"github.com/gin-gonic/gin"
	"galaxyotc/common/middleware"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("im_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 搜索用户信息
		api.GET("/search", controller.Search)
		// 拉取聊天消息
		api.GET("/history", middleware.SigninRequired, controller.History)
		// 接收消息抄送
		api.POST("/receiveMsg", controller.ReceiveMsg)
	}
}
