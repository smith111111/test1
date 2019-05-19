package router

import (
	"galaxyotc/common/middleware"

	"galaxyotc/gc_services/otc_service/controller/order"
	"galaxyotc/gc_services/otc_service/controller/offer"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("otc_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// OTC广告
		{
			// 发布新广告
			api.POST("/offer", middleware.SigninRequired, offer.NewOffer)
			// 编辑广告
			api.PUT("/offer", middleware.SigninRequired, offer.UpdateOffer)
			// 获取指定的类型的广告列表
			api.GET("/offers/:type", offer.AllOffers)
			// 获取指定的广告详情
			api.GET("/offer/:code", middleware.SetContextUser, offer.OfferDetail)
			// 获取我的广告列表
			api.GET("/my/offers", middleware.SigninRequired, offer.MyAllOffers)
			// 广告上下架
			api.POST("/make/offer", middleware.SigninRequired, offer.MakeOffer)
		}

		// OTC订单
		{
			// 获取最新交易记录
			api.GET("/orders/recently", order.RecentlyOrders)
			// 根据筛选条件获取订单列表
			api.GET("/orders", middleware.SigninRequired, order.Orders)
			// 创建新订单
			api.POST("/order", middleware.SigninRequired, order.NewOrder)
			// 根据订单id查询明细
			api.GET("/order/:sn", middleware.SigninRequired, order.OrderDetail)
			// 卖家同意接单
			api.POST("/order/approved", middleware.SigninRequired, order.Approved)
			// 买家确认付款
			api.POST("/order/completed/pay", middleware.SigninRequired, order.CompletedPay)
			// 卖家同意放币
			api.POST("/order/release", middleware.SigninRequired, order.Release)
			// 买家确认收币
			api.POST("/order/completed", middleware.SigninRequired, order.Completed)
			// 取消订单
			api.POST("/order/canceled", middleware.SigninRequired, order.Canceled)
		}
	}
}
