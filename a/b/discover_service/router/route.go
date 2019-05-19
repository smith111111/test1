package router

import (
	"galaxyotc/common/middleware"

	"galaxyotc/gc_services/discover_service/controller/menuProduct"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("discover_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		{
			api.GET("/type/menuProduct",menuProduct.MenuProduct)
			// 获取菜单产品
			//api.GET("/type/menuProduct/:menu_id",menuProduct.MenuProduct)
			// 获取菜单产品
			api.GET("/type/menuList",menuProduct.MenuList)

			api.GET("/type/findmenuProduct",menuProduct.FindmenuProduct)
		}

	}
}
