package router

import (
	"github.com/gin-gonic/gin"
	"galaxyotc/common/middleware"
	"galaxyotc/gc_services/user_service/controller/user"
	"galaxyotc/gc_services/user_service/controller/verify"
	"galaxyotc/gc_services/user_service/controller/trading_method"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("user_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 用户注册
		api.POST("/signup", user.Signup)
		// 用户登录
		api.POST("/signin", user.Signin)
		// 注销
		api.POST("/signout", middleware.SigninRequired, user.Signout)
		// 用户个人信息
		api.GET("/info", middleware.SigninRequired, user.Info)
		// 绑定手机或者邮箱
		api.POST("/bind", middleware.SigninRequired, user.BindMobileOrEmail)
		// 校验密码
		api.POST("/password/check", middleware.SigninRequired, user.CheckPassword)
		// 修改密码
		api.POST("/password/reset", middleware.SigninRequired, user.ResetPassword)
		// 找回密码
		api.POST("/password/get_back", user.ResetPasswordForGetBack)
		// 修改用户头像
		api.POST("/avatar", middleware.SigninRequired, user.UpdateAvatar)
		// 修改用户姓名
		api.POST("/name", middleware.SigninRequired, user.UpdateName)
		// 获取我的邀请会员
		api.GET("/invitees", middleware.SigninRequired, user.UserInvitees)
		// 获取我的收益
		api.GET("/commissions", middleware.SigninRequired, user.UserCommissions)

		// 实名认证
		api.POST("/verify", middleware.SigninRequired, verify.Verify)

		// 添加支付方式信息
		api.POST("/trading_method/create", middleware.SigninRequired, trading_method.Create)
		// 修改支付方式信息
		api.POST("/trading_method/update", middleware.SigninRequired, trading_method.Update)
		// 删除支付方式信息
		api.GET("/trading_method/:id", middleware.SigninRequired, trading_method.DeleteTradingMethod)
		// 获取用户交易方式列表
		api.GET("/trading_methods", middleware.SigninRequired, trading_method.UserTradingMethods)
	}
}
