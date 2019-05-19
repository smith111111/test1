package router

import (
	"github.com/gin-gonic/gin"
	"galaxyotc/common/middleware"
	"galaxyotc/gc_services/captcha_service/controller"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("captcha_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 获取验证码
		api.POST("/user", controller.SendCaptchaByID)
		// 获取验证码（注册）
		api.POST("/signup", controller.SendSignupCaptcha)
		// 获取验证码（找回密码）
		api.POST("/forgot", controller.SendForgotCaptcha)
		// 获取手机验证码（注册，带人机验证）
		//api.POST("/web/captcha/mobile", captcha.SendSignupMobileCaptchaForWeb)
	}
}
