package currency

import (
	"math"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/utils"
	"galaxyotc/common/model"

	"galaxyotc/gc_services/base_service/api"
)

// 获取代币信息
func CryptoCurrency(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	code := c.Param("code")
	if code == "" {
		SendErrJSON("币种不能为空", c)
		return
	}

	// 获取币种信息
	currency, err := model.CurrencyFromAndToRedis(code)
	if err != nil {
		log.Errorf("Currency-CryptoCurrency-Error: %s", err.Error())
		SendErrJSON("获取币种信息失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currency": currency,
		},
	})
}

// 获取法币信息
func FiatCurrency(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	code := c.Param("code")
	if code == "" {
		SendErrJSON("币种不能为空", c)
		return
	}

	// 获取币种信息
	currency, err := model.FiatCurrencyFromAndToRedis(code)
	if err != nil {
		log.Errorf("Currency-FiatCurrency-Error: %s", err.Error())
		SendErrJSON("获取币种信息失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currency": currency,
		},
	})
}

// 币种基本信息
type currencyBaseInfo struct {
	Name      string `json:"name"`
	Code      string `json:"code"`
	Symbol    string `json:"symbol"`
	Icon      string `json:"icon"`
	Precision uint   `json:"precision"`
}

// 代币列表
func CryptoCurrencies(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var currencies []*currencyBaseInfo

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var totalCount int64

	baseQuery := model.DB.Table("currencies").Where("status = ?", model.CurrencyNormalInt)

	if err := baseQuery.Offset(offset).Limit(size).Find(&currencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-CryptoCurrencies-Error: %s", err.Error())
		SendErrJSON("获取代币列表失败", c)
		return
	}

	// 获取代币总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Currency-CryptoCurrencies-Error: %s", err.Error())
		SendErrJSON("获取代币总数量失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currencies": currencies,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 普通用户法币列表
func FiatCurrencies(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var currencies []*currencyBaseInfo

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var totalCount int64

	baseQuery := model.DB.Table("fiat_currencies").Where("status = ?", model.CurrencyNormalInt)

	if err := baseQuery.Offset(offset).Limit(size).Find(&currencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-FiatCurrencies-Error: %s", err.Error())
		SendErrJSON("获取法币列表失败", c)
		return
	}

	// 获取代币总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Currency-FiatCurrencies-Error: %s", err.Error())
		SendErrJSON("获取法币总数量失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currencies": currencies,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 代币和法币列表
func CryptoAndFiatCurrencies(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	currencies := []*currencyBaseInfo{}
	fiatCurrencies := []*currencyBaseInfo{}

	// 获取代币信息
	if err := model.DB.Table("currencies").Where("status = ?", model.CurrencyNormalInt).Find(&currencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-CryptoAndFiatCurrencies-Error: %s", err.Error())
		SendErrJSON("获取代币列表失败", c)
		return
	}

	// 获取法币信息
	if err := model.DB.Table("fiat_currencies").Where("status = ?", model.CurrencyNormalInt).Find(&fiatCurrencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-CryptoAndFiatCurrencies-Error: %s", err.Error())
		SendErrJSON("获取法币列表失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currency":      currencies,
			"fiat_currency": fiatCurrencies,
		},
	})
}

type exchangerateInfo struct {
	Symbol 	string		`json:"symbol"`
	Rate	float64		`json:"rate"`
}


// 获取代币的所有法币汇率
func Exchangerates(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	ExchangeratesMap := make(map[string]map[string]*exchangerateInfo)

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var (
		currencies []*model.Currency
		fiatCurrencies []*model.FiatCurrency
	)

	baseQuery := model.DB.Model(model.Currency{})

	if err := baseQuery.Where("status = ?", model.CurrencyNormalInt).Offset(offset).Limit(size).Find(&currencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-Exchangerates-Error: %s", err.Error())
		SendErrJSON("获取代币列表失败", c)
		return
	}

	// 获取所有的法币信息，不需要分页
	if err := model.DB.Model(model.FiatCurrency{}).Where("status = ?", model.CurrencyNormalInt).Find(&fiatCurrencies).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Currency-Exchangerates-Error: %s", err.Error())
		SendErrJSON("获取法币列表失败", c)
		return
	}

	var totalCount int64

	// 获取代币总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Currency-Exchangerates-Error: %s", err.Error())
		SendErrJSON("获取代币总数量失败", c)
		return
	}

	if totalCount > 0 {
		// 获取实时汇率
		allRates, err := api.ExchangerateApi.GetAllRates(true)
		if err != nil {
			log.Errorf("Currency-Exchangerates-Error: %s", err.Error())
			SendErrJSON("获取代币实时汇率失败", c)
			return
		}

		for _, c := range currencies {

			// 获取代币的当前汇率
			cryptoRate := decimal.NewFromFloat(allRates[c.Code])
			currencyRates := make(map[string]*exchangerateInfo)

			if len(fiatCurrencies) > 0 {
				for _, f := range fiatCurrencies {
					// 获取法币的当前汇率
					fiatRate := decimal.NewFromFloat(allRates[f.Code])
					exchangeRate, _ :=  fiatRate.DivRound(cryptoRate, 2).Float64()

					if utils.IsNaNOrInf(exchangeRate) {
						exchangeRate = 0
					}

					currencyRates[f.Code] = &exchangerateInfo{f.Symbol, exchangeRate}
				}
			}

			ExchangeratesMap[c.Code] = currencyRates
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"currencies": ExchangeratesMap,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 获取指定代币的所有法币汇率
func CurrencyExchangerates(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	code := strings.ToUpper(c.Param("code"))

	var fiatCurrencies []*model.FiatCurrency
	if err := model.DB.Model(model.FiatCurrency{}).Where("status = ?", model.CurrencyNormalInt).Find(&fiatCurrencies).Error; err != nil {
		log.Errorf("Currency-CurrencyExchangerates-Error: %s", err.Error())
		SendErrJSON("获取代币列表失败", c)
		return
	}

	// 获取实时汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Currency-CurrencyExchangerates-Error: %s", err.Error())
		SendErrJSON("获取代币实时汇率失败", c)
		return
	}

	ExchangeratesMap := make(map[string]*exchangerateInfo)

	for _, fiat := range fiatCurrencies {
		// 单个币种对应法币的价值
		fiatRate := decimal.NewFromFloat(allRates[fiat.Code])
		cryptoRate := decimal.NewFromFloat(allRates[code])
		exchangeRate, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

		if utils.IsNaNOrInf(exchangeRate) {
			exchangeRate = 0
		}

		ExchangeratesMap[fiat.Code] = &exchangerateInfo{fiat.Symbol, exchangeRate}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  ExchangeratesMap,
	})
}
