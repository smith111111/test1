package router

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"galaxyotc/common/middleware"

	"galaxyotc/gc_services/base_service/controller/base"
	"galaxyotc/gc_services/base_service/controller/area"
	"galaxyotc/gc_services/base_service/controller/trading_method"
	"galaxyotc/gc_services/base_service/controller/trading_rates"
	"galaxyotc/gc_services/base_service/controller/notice"
	"galaxyotc/gc_services/base_service/controller/currency"
)

// Route 路由
func Route(router *gin.Engine) {
	apiPrefix := viper.GetString("server.api_prefix") + viper.GetString("base_service.api_prefix")
	api := router.Group(apiPrefix, middleware.RefreshTokenCookie)
	{
		// 获取App版本
		api.GET("/version", base.Version)
		// 获取国家地区列表
		api.GET("/areas", area.Areas)
		// 获取交易方式列表
		api.GET("/trading_methods", trading_method.TradingMethods)
		// 获取交易汇率列表
		api.GET("/trading_rates", trading_rates.TradingRates)
		// 获取公告列表
		api.GET("/notices", notice.Notices)
		// 获取公告详情
		api.GET("/notice/:id", notice.NoticeDetail)
		// 分发佣金
		//api.GET("/distributions", commission_distribution.DistributionTest)

		// 币种信息
		currencyApi := api.Group("currency")
		{
			// 获取有效的代币信息和有效的法币信息
			currencyApi.GET("/all", currency.CryptoAndFiatCurrencies)
			// 获取所有代币信息
			currencyApi.GET("/cryptos", currency.CryptoCurrencies)
			// 获取所有代币信息
			currencyApi.GET("/fiats", currency.FiatCurrencies)
			// 获取指定代币信息
			currencyApi.GET("/crypto/:code", currency.CryptoCurrency)
			// 获取指定法币信息
			currencyApi.GET("/fiat/:code", currency.FiatCurrency)
			// 获取代币对应的所有法币的汇率
			currencyApi.GET("/exchange_rates", currency.Exchangerates)
			// 获取指定代币对应的所有法币的汇率
			currencyApi.GET("/exchange_rates/:code", currency.CurrencyExchangerates)
		}
	}
}
