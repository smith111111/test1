package trading_rates

import (
	"math"
	"net/http"
	"strconv"
	"galaxyotc/common/log"
	"galaxyotc/common/net"
	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Create 创建交易费率
func CreateTradingRate(c *gin.Context) {
	saveTradingRate(c, true)
}

// Update 更新交易费率
func UpdateTradingRate(c *gin.Context) {
	saveTradingRate(c, false)
}

func saveTradingRate (c *gin.Context, isCrate bool) {
	SendErrJSON := net.SendErrJSON
	var tradingRate model.TradingRate
	if err := c.ShouldBindJSON(&tradingRate); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	// 检查是否缺少必要的参数
	if tradingRate.TradeType < 0 {
		SendErrJSON("交易类型参数出错", c)
		return
	}

	if isCrate {
		if !model.DB.Where("trade_type = ? and trade_mode = ? and coin_code = ?", tradingRate.TradeType, tradingRate.TradeMode, tradingRate.CoinCode).NewRecord(&tradingRate) {
			SendErrJSON("当前已经存在相同的数据", c)
			return
		}

		if err := model.DB.Create(&tradingRate).Error; err != nil {
			log.Errorf("SaveTradingRate Error: %s", err.Error())
			SendErrJSON("新增交易费率失败", c)
			return
		}
	} else {
		var updateTradingRate model.TradingRate
		if err := model.DB.First(&updateTradingRate, tradingRate.ID).Error; err == nil {
			updateTradingRate.Rate = tradingRate.Rate
			updateTradingRate.MinFee = tradingRate.MinFee
			updateTradingRate.MaxFee = tradingRate.MaxFee
			updateTradingRate.FeeCoinCode = tradingRate.FeeCoinCode
			updateTradingRate.MinAmount = tradingRate.MinAmount
			updateTradingRate.MaxAmount = tradingRate.MaxAmount
			if err := model.DB.Save(&updateTradingRate).Error; err != nil {
				log.Errorf("SaveTradingRate Error: %s", err.Error())
				SendErrJSON("更新交易费率失败", c)
				return
			}
		} else {
			SendErrJSON("无效的交易费率", c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// Delete 删除交易费率
func DeleteTradingRate(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var id int
	var idErr error
	if id, idErr = strconv.Atoi(c.Param("id")); idErr != nil {
		SendErrJSON("无效的id", c)
		return
	}

	var tradingRate model.TradingRate

	if err := model.DB.First(&tradingRate, id).Error; err != nil {
		SendErrJSON("无效的id", c)
		return
	}

	if err := model.DB.Delete(&tradingRate).Error; err != nil {
		log.Errorf("DeleteTradingRate Error: %s", err.Error())
		SendErrJSON("error", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"id": id,
		},
	})
}

// 获取交易费率列表
func TradingRates(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	//交易类型
	tradeType := c.Query("type")

	//交易模式
	tradeMode := c.Query("mode")

	//币种
	code := c.Query("code")

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var (
		tradingRates []*model.TradingRate
		totalCount       int64
	)

	var (
		selectSQL string
		args      []interface{}
	)

	// 是否进行交易类型筛选
	if code != "" {
		selectSQL += " coin_code = ?"
		args = append(args, code)
	}

	// 是否进行交易类型筛选
	if tradeType != "" {
		selectSQL += " and trade_type = ?"
		args = append(args, tradeType)
	}

	// 是否进行交易模式筛选
	if tradeMode != "" {
		selectSQL += " and trade_mode = ?"
		args = append(args, tradeMode)
	}

	baseQuery := model.DB.Model(&model.TradingRate{}).Where(selectSQL, args...).Order("coin_code")

	if err := baseQuery.Offset(offset).Limit(size).Find(&tradingRates).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("TradingRates Error: %s", err.Error())
		SendErrJSON("获取交易费率列表失败", c)
		return
	}

	// 获取交易方式总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("TradingRates Error: %s", err.Error())
		SendErrJSON("获取交易费率列表失败", c)
		return
	}

	tradingRateList := model.TradingRateInfoList{}

	if len(tradingRates) > 0 {
		for _, item := range tradingRates {
			offerInfo := model.TradingRateInfo{
				ID:          item.ID,
				TradeType:   item.TradeType,
				TradeMode:   item.TradeMode,
				CoinCode:    item.CoinCode,
				Rate:        item.Rate,
				MinFee:      item.MinFee,
				MaxFee:      item.MaxFee,
				FeeCoinCode: item.FeeCoinCode,
				MinAmount:   item.MinAmount,
				MaxAmount:   item.MaxAmount,
			}

			// 根据枚举获取对应详情
			offerInfo.TradeTypeString = model.TradeTypeString(item.TradeType)
			offerInfo.TradeModeString = model.TradeModeString(item.TradeMode)

			tradingRateList = append(tradingRateList, &offerInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"tradingRates": 	tradingRateList,
			"pageNo":          page,
			"pageSize":        size,
			"totalPage":       math.Ceil(float64(totalCount) / float64(size)),
			"totalCount":      totalCount,
		},
	})
}
