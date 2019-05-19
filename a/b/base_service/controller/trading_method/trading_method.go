package trading_method

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// 保存交易方式信息
func saveTradingMethod(c *gin.Context, isCreate bool) {
	SendErrJSON := net.SendErrJSON
	var tradingMethod model.TradingMethod
	if err := c.ShouldBindJSON(&tradingMethod); err != nil {
		fmt.Println(err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if tradingMethod.Name == "" {
		SendErrJSON("交易方式名称不能为空", c)
		return
	}

	if isCreate {
		if !model.DB.First(&tradingMethod, &model.TradingMethod{Name: tradingMethod.Name}).RecordNotFound() {
			SendErrJSON("交易方式信息已存在", c)
			return
		}

		if err := model.DB.Create(&tradingMethod).Error; err != nil {
			SendErrJSON("保存交易方式信息失败", c)
			return
		}
	} else {
		var updateTradingMethod model.TradingMethod
		if err := model.DB.First(&updateTradingMethod, tradingMethod.ID); err != nil {
			SendErrJSON("无效的交易方式", c)
			return
		}

		if err := model.DB.Model(&updateTradingMethod).Updates(model.TradingMethod{Name: tradingMethod.Name, EnName: tradingMethod.EnName, Icon: tradingMethod.Icon}).Error; err != nil {
			SendErrJSON("更新交易方式信息失败", c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{},
	})
}

// Create 创建交易方式信息
func CreateTradingMethod(c *gin.Context) {
	saveTradingMethod(c, true)
}

// Update 更新交易方式信息
func UpdateTradingMethod(c *gin.Context) {
	saveTradingMethod(c, false)
}

// 禁用交易方式信息
func DisableTradingMethod(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var (
		ID 	int64
		err error
	)
	if ID, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
		SendErrJSON("交易方式ID有误", c)
		return
	}

	if err := model.DB.Model(&model.TradingMethod{}).Where("id = ?", ID).Update("is_deleted", true).Error; err != nil {
		log.Errorf("DisableTradingMethod Error: %s", err.Error())
		SendErrJSON("禁用交易方式失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"ID": ID,
		},
	})
}

// Recover 恢复交易方式信息
func Recover(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var (
		ID 	int64
		err error
	)

	if ID, err = strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
		SendErrJSON("交易方式ID有误", c)
		return
	}

	if err := model.DB.Model(&model.TradingMethod{}).Where("id = ?", ID).Update("is_deleted", false).Error; err != nil {
		log.Errorf("DisableTradingMethod Error: %s", err.Error())
		SendErrJSON("恢复交易方式失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"ID": ID,
		},
	})
}

// 获取交易方式列表
func queryTradingMethods(c *gin.Context, isBackend bool) {
	SendErrJSON := net.SendErrJSON

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var (
		tradingMethods []*model.TradingMethod
		totalCount int64
	)

	var baseQuery 	*gorm.DB

	// 后台管理查询交易方式时，会返回已禁用的交易方式
	if isBackend {
		baseQuery = model.DB.Model(model.TradingMethod{})
	} else {
		baseQuery = model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false)
	}

	if err := baseQuery.Offset(offset).Limit(size).Find(&tradingMethods).Error; err != nil && err != gorm.ErrRecordNotFound {
		SendErrJSON("获取交易方式列表失败", c)
		return
	}

	// 获取交易方式总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		fmt.Println(err.Error())
		SendErrJSON("获取交易方式列表失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"trading_methods": tradingMethods,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 普通用户交易方式列表
func TradingMethods(c *gin.Context) {
	queryTradingMethods(c, false)
}

// 后台管理交易方式列表
func AllTradingMethods(c *gin.Context) {
	queryTradingMethods(c, true)
}