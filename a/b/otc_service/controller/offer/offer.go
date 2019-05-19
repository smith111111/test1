package offer

import (
	"fmt"
	"math"
	"time"
	"sort"
	"strings"
	"net/http"

	"galaxyotc/common/log"
	"galaxyotc/common/net"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"

	"galaxyotc/gc_services/otc_service/api"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/garyburd/redigo/redis"
)


// 保存广告（创建或更新）
func saveOffer(c *gin.Context, isCreate bool) {
	SendErrJSON := net.SendErrJSON
	var offer model.Offer
	if err := c.ShouldBindJSON(&offer); err != nil {
		log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	//校验用户是否进行了实名验证
	/*if !user.IsRealName {
		SendErrJSON("请先进行实名验证", c)
		return
	}*/

	switch offer.OfferType {
	case model.BuyOfferInt:
	case model.SellOfferInt:
	default:
		SendErrJSON("无效的广告类型", c)
		return
	}

	switch offer.TradingType {
	case model.WalletTradingInt:
	case model.AccountTradingInt:
	default:
		SendErrJSON("无效的交易类型", c)
		return
	}

	// 检查是否缺少必要的参数
	if offer.TradingMethods == "" || offer.Price == 0 || offer.MinLimit == 0 || offer.MaxLimit == 0 {
		SendErrJSON("缺少必要的参数", c)
		return
	}

	// 根据代币代码获取代币信息
	currency, err := model.CurrencyFromAndToRedis(offer.Currency)
	if err != nil {
		log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 根据法币代码获取法币信息
	fiatCurrency, err := model.FiatCurrencyFromAndToRedis(offer.FiatCurrency)
	if err != nil {
		log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
		SendErrJSON("无效的法币类型", c)
		return
	}

	// 获取Redis连接
	RedisConn := model.RedisPool.Get()
	defer RedisConn.Close()

	limitKey := fmt.Sprintf("%s%s%s%d", model.PushOfferLimit, currency.Code, fiatCurrency.Code, user.ID)
	limitCount, err := redis.Int64(RedisConn.Do("GET", limitKey))
	if err == nil && limitCount >= model.PushOfferLimitCount {
		log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
		SendErrJSON("无法发布广告，您已超过发布广告的限额", c)
		return
	}

	if isCreate {
		offer.Currency = currency.Code
		offer.FiatCurrency = fiatCurrency.Code
		offer.AccountID = user.ID
		// 根据当前时间生成随机码
		offer.Code = xid.NewWithTime(time.Now().Local()).String()

		if err := model.DB.Create(&offer).Error; err != nil {
			log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
			SendErrJSON("发布新广告失败", c)
			return
		}

		// 发布广告成功后添加用户广告次数
		if _, err := RedisConn.Do("SET", limitKey, limitCount + 1); err != nil {
			log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
			SendErrJSON("服务器出错啦", c)
			return
		}
	} else {
		var updatedOffer model.Offer
		if err := model.DB.Where("code = ?", offer.Code).First(&updatedOffer).Error; err == nil {
			if updatedOffer.AccountID != user.ID {
				SendErrJSON("您没有权限执行此操作", c)
				return
			}

			updatedOffer.Currency = currency.Code
			updatedOffer.FiatCurrency = fiatCurrency.Code
			updatedOffer.OfferType = offer.OfferType
			updatedOffer.TradingType = offer.TradingType
			updatedOffer.Float = offer.Float
			updatedOffer.AcceptPrice = offer.AcceptPrice
			updatedOffer.Price = offer.Price
			updatedOffer.TradingMethods = offer.TradingMethods
			updatedOffer.MinLimit = offer.MinLimit
			updatedOffer.MaxLimit = offer.MaxLimit
			updatedOffer.Note = offer.Note
			if err := model.DB.Save(&updatedOffer).Error; err != nil {
				log.Errorf("Offer-SaveOffer-Error: %s", err.Error())
				SendErrJSON("更新广告失败", c)
				return
			}
		} else {
			SendErrJSON("无效的广告", c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// Create 创建广告
func NewOffer(c *gin.Context) {
	saveOffer(c, true)
}

// Update 更新广告
func UpdateOffer(c *gin.Context) {
	saveOffer(c, false)
}

// 获取指定广告类型的广告列表
func AllOffers(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取广告类型，默认为出售
	offerTypeStr := c.Param("type")
	if offerTypeStr == "" {
		SendErrJSON("获取广告类型失败", c)
		return
	}

	var offerType int32
	switch offerTypeStr {
	// 选择购买返回出售类型广告
	case "buy":
		offerType = model.SellOfferInt
		// 选择出售返回购买类型广告
	case "sell":
		offerType = model.BuyOfferInt
	}

	// 获取代币类型，默认为BTC
	currencyCode := strings.ToUpper(c.DefaultQuery("currency", "btc"))

	// 获取法币类型，默认为CNY
	fiatCurrencyCode := strings.ToUpper(c.DefaultQuery("fiat_currency", "cny"))

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var tradingMethod []*model.TradingMethod
	// 获取所有的交易方式
	if err := model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false).Find(&tradingMethod).Error; err != nil {
		log.Errorf("Offer-AllOffers-Error: %s", err.Error())
		SendErrJSON("获取交易方式信息失败", c)
		return
	}

	tradingMethodMap := make(map[uint64]*model.TradingMethod)
	if len(tradingMethod) > 0 {
		for _, method := range tradingMethod {
			tradingMethodMap[method.ID] = method
		}
	}

	var (
		offers     []*model.Offer
		totalCount int64
	)

	baseQuery := model.DB.Model(&model.Offer{}).Where(&model.Offer{OfferType: offerType, Currency: currencyCode, FiatCurrency: fiatCurrencyCode, Status: model.OfferOnInt}).Order("price")

	if err := baseQuery.Offset(offset).Limit(size).Find(&offers).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Offer-AllOffers-Error: %s", err.Error())
		SendErrJSON("获取广告列表失败", c)
		return
	}

	// 获取广告总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Offer-AllOffers-Error: %s", err.Error())
		SendErrJSON("获取广告列表失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Offer-AllOffers-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 获取当前的浮动汇率
	fiatRate := decimal.NewFromFloat(allRates[fiatCurrencyCode])
	cryptoRate := decimal.NewFromFloat(allRates[currencyCode])
	currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
	if utils.IsNaNOrInf(currentPrice) {
		currentPrice = 0
	}

	offerList := model.OfferList{}
	if len(offers) > 0 {
		for _, offer := range offers {
			var offerInfo model.OfferInfo
			if err := offerInfo.Init(currentPrice, offer, tradingMethodMap); err != nil {
				SendErrJSON(err.Error(), c)
				return
			}

			offerList = append(offerList, &offerInfo)
		}
	}

	// 对广告列表进行排序
	sort.Sort(offerList)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"offers":     offerList,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 获取指定广告信息
func OfferDetail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取广告类型，默认为出售
	code := c.Param("code")
	if code == "" {
		SendErrJSON("广告码不能为空", c)
		return
	}

	var tradingMethod []*model.TradingMethod
	// 获取所有支持的交易方式
	if err := model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false).Find(&tradingMethod).Error; err != nil {
		log.Errorf("Offer-OfferDetail-Error: %s", err.Error())
		SendErrJSON("获取交易方式信息失败", c)
		return
	}

	tradingMethodMap := make(map[uint64]*model.TradingMethod)
	if len(tradingMethod) > 0 {
		for _, method := range tradingMethod {
			tradingMethodMap[method.ID] = method
		}
	}

	var offer model.Offer

	err := model.DB.Model(&model.Offer{}).Where("code = ?", code).First(&offer).Error

	if err == gorm.ErrRecordNotFound {
		SendErrJSON("广告不存在", c)
		return
	} else if err != nil {
		log.Errorf("Offer-OfferDetail-Error: %s", err.Error())
		SendErrJSON("获取广告详情失败", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	if userInter != nil {
		user := userInter.(model.User)
		// 判断当前用户是否是广告卖家
		if user.ID == offer.AccountID {
			SendErrJSON("无法查看自己发布的广告", c)
			return
		}
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Offer-OfferDetail-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 获取当前的浮动汇率
	fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
	cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
	currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
	if utils.IsNaNOrInf(currentPrice) {
		currentPrice = 0
	}

	var offerInfo model.OfferInfo
	if err := offerInfo.Init(currentPrice, &offer, tradingMethodMap); err != nil {
		log.Errorf("Offer-OfferDetail-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"offer": offerInfo,
		},
	})
}

// 获取我的广告列表
func MyAllOffers(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取广告类型,默认是所有,on: 上货,off: 下架
	status := c.DefaultQuery("status", "all")

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var tradingMethod []*model.TradingMethod
	// 获取所有的交易方式
	if err := model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false).Find(&tradingMethod).Error; err != nil {
		log.Errorf("Offer-MyAllOffers-Error: %s", err.Error())
		SendErrJSON("获取交易方式信息失败", c)
		return
	}

	tradingMethodMap := make(map[uint64]*model.TradingMethod)
	if len(tradingMethod) > 0 {
		for _, method := range tradingMethod {
			tradingMethodMap[method.ID] = method
		}
	}

	var (
		offers     []*model.Offer
		totalCount int64
	)

	var (
		selectSQL string
		args      []interface{}
	)

	selectSQL = "account_id = ?"
	args = append(args, user.ID)

	// 是否进行状态筛选
	switch status {
	case "all":
		// 不处理
	case "on":
		selectSQL += " and status = ?"
		args = append(args, model.OfferOnInt)
	case "off":
		selectSQL += " and status = ?"
		args = append(args, model.OfferOffInt)
	}

	baseQuery := model.DB.Model(&model.Offer{}).Where(selectSQL, args...).Order("price")

	if err := baseQuery.Offset(offset).Limit(size).Find(&offers).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Offer-MyAllOffers-Error: %s", err.Error())
		SendErrJSON("获取广告列表失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Offer-MyAllOffers-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	offerList := model.OfferList{}

	if len(offers) > 0 {
		for _, offer := range offers {
			// 获取当前的浮动汇率
			fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
			cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
			currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

			var offerInfo model.OfferInfo
			if err := offerInfo.Init(currentPrice, offer, tradingMethodMap); err != nil {
				log.Errorf("Offer-MyAllOffers-Error: %s", err.Error())
				SendErrJSON(err.Error(), c)
				return
			}

			offerList = append(offerList, &offerInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"offers":     offerList,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 广告上下架
func MakeOffer(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	type ReqData struct {
		Code   string `json:"code"`
		Status int32    `json:"status"`
	}
	var repData ReqData
	if err := c.ShouldBindWith(&repData, binding.JSON); err != nil {
		log.Errorf("Offer-MakeOffer-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var updateOffer model.Offer
	if err := model.DB.Where("code = ?", repData.Code).First(&updateOffer).Error; err != nil {
		log.Errorf("Offer-MakeOffer-Error: %s", err.Error())
		SendErrJSON("error", c)
		return
	}

	if updateOffer.Status == repData.Status && repData.Status == model.OfferOffInt {
		SendErrJSON("已下架的广告不能再下架", c)
		return
	} else if updateOffer.Status == repData.Status && repData.Status == model.OfferOnInt {
		SendErrJSON("已上架的广告不能再上架", c)
		return
	}

	if err := model.DB.Model(&updateOffer).Update("status", repData.Status).Error; err != nil {
		log.Errorf("Offer-MakeOffer-Error: %s", err.Error())
		SendErrJSON("操作失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{},
	})
}