package router

import (
	"github.com/gin-gonic/gin"
	"galaxyotc/common/middleware"
	"galaxyotc/gc_services/account_service/controller/wallet"
	"galaxyotc/gc_services/account_service/controller/deposit"
	"galaxyotc/gc_services/account_service/controller/withdraw"
	"github.com/spf13/viper"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("account_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 资产管理
		api.GET("/wallets", middleware.SigninRequired, wallet.Wallets)
		// 根据指定币种的钱包信息
		api.GET("/wallet", middleware.SigninRequired, wallet.CurrencyWallet)
		// 数据统计
		api.GET("/statistics", middleware.SigninRequired, wallet.Statistics)
		// 获取充值/提币记录
		api.GET("/histories", middleware.SigninRequired, wallet.Histories)
		// 交易转账
		api.POST("/transfer", middleware.SigninRequired, wallet.Transfer)
		// 交易转账记录
		api.GET("/transfers", middleware.SigninRequired, wallet.Transfers)

		// 账户充值
		depositApi := api.Group("deposit")
		{
			// 获取指定币种的充值地址
			depositApi.GET("/address", middleware.SigninRequired, deposit.DepositAddress)
			// 获取充值交易记录详情
			depositApi.GET("/detail/:sn", middleware.SigninRequired, deposit.DepositDetail)
		}

		// 账户提现
		withdrawApi := api.Group("withdraw")
		{
			// 提币
			withdrawApi.POST("", middleware.SigninRequired, withdraw.Withdraw)
			// 获取提币交易记录详情
			withdrawApi.GET("/detail/:sn", middleware.SigninRequired, withdraw.WithdrawDetail)
		}
	}
}
