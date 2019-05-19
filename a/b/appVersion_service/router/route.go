package router

import (
	"galaxyotc/common/middleware"
	"galaxyotc/gc_services/appVersion_service/controller/appVersionProduct"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("appVersion_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		{
			//获取最新版本
			api.GET("/type/lastAppVersion",appVersionProduct.AppVersionProduct)
		}

	}
}
