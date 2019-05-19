package trading_method

import (
	"math"
	"errors"
	"strconv"
	"strings"
	"net/http"

	"galaxyotc/common/net"
	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mozillazg/go-pinyin"
	"galaxyotc/common/log"
)

// 保存用户支付方式信息
func saveTradingMethod(c *gin.Context, isCreate bool) {
	SendErrJSON := net.SendErrJSON

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var tradingMethod model.UserTradingMethod
	if err := c.ShouldBindJSON(&tradingMethod); err != nil {
		log.Errorf("User-SaveTradingMethod-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if tradingMethod.AccountNumber == "" {
		SendErrJSON("账号不能为空", c)
		return
	}

	if tradingMethod.Payee == "" {
		SendErrJSON("收款人不能为空", c)
		return
	}

	if tradingMethod.TradingMethodID <= 0 {
		SendErrJSON("交易方式不正确", c)
		return
	}

	if model.DB.Model(&model.TradingMethod{}).Where("id = ? and is_deleted = ?", tradingMethod.TradingMethodID, false).RecordNotFound() {
		SendErrJSON("交易方式无效", c)
		return
	}

	if isCreate {
		tradingMethod.UserID = user.ID
		tradingMethod.PayeePinyin = strings.Join(pinyin.LazyConvert(tradingMethod.Payee, nil), "")

		if err := model.DB.Create(&tradingMethod).Error; err != nil {
			log.Errorf("User-SaveTradingMethod-Error: %s", err.Error())
			SendErrJSON("添加支付方式失败", c)
			return
		}
	} else {
		var updateTradingMethod model.UserTradingMethod
		if err := model.DB.First(&updateTradingMethod, tradingMethod.ID).Error; err != nil || updateTradingMethod.IsDeleted {
			SendErrJSON("无效的支付方式", c)
			return
		}

		if updateTradingMethod.UserID != user.ID {
			SendErrJSON("没有权限执行此操作", c)
			return
		}

		UpdatesInfo := make(map[string]interface{})

		if updateTradingMethod.BankName != "" {
			UpdatesInfo["bank_name"] = tradingMethod.BankName
		}

		if updateTradingMethod.DepositBank != "" {
			UpdatesInfo["deposit_bank"] = tradingMethod.DepositBank
		}

		if updateTradingMethod.Payee != "" {
			UpdatesInfo["payee"] = tradingMethod.Payee
			UpdatesInfo["payee_pinyin"] = strings.Join(pinyin.LazyConvert(tradingMethod.Payee, nil), "")
		}

		if updateTradingMethod.AccountNumber != "" {
			UpdatesInfo["account_number"] = tradingMethod.AccountNumber
		}

		if err := model.DB.Model(&updateTradingMethod).Save(&UpdatesInfo).Error; err != nil {
			log.Errorf("User-SaveTradingMethod-Error: %s", err.Error())
			SendErrJSON("更新支付方式失败", c)
			return
		}
	}

	// 保存成功后，直接返回最新列表
	page, size := net.GetPageAndSize(c)
	resp, err := getByUser(user.ID, page, size)
	if err != nil {
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": resp,
	})
}

// 添加用户支付方式信息
func Create(c *gin.Context) {
	saveTradingMethod(c, true)
}

// 修改用户支付方式信息
func Update(c *gin.Context) {
	saveTradingMethod(c, false)
}

// 删除用户支付方式信息
func DeleteTradingMethod(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		SendErrJSON("无效的参数", c)
		return
	}

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var userTradingMethod model.UserTradingMethod
	if err := model.DB.Where("id = ? and is_deleted = ?", id, false).First(&userTradingMethod).Error; err != nil {
		log.Errorf("User-DeleteTradingMethod-Error: %s", err.Error())
		SendErrJSON("无效的支付方式或已删除", c)
		return
	}

	if userTradingMethod.UserID != user.ID {
		SendErrJSON("没有权限执行此操作", c)
		return
	}

	userTradingMethod.IsDeleted = true;
	if err := model.DB.Save(&userTradingMethod).Error; err != nil {
		log.Errorf("User-DeleteTradingMethod-Error: %s", err.Error())
		SendErrJSON("删除支付方式失败", c)
		return
	}

	// 删除成功后，直接返回最新列表
	page, size := net.GetPageAndSize(c)
	resp, err := getByUser(user.ID, page, size)
	if err != nil {
		log.Errorf("User-DeleteTradingMethod-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": resp,
	})
}

// 获取用户交易方式信息列表
func UserTradingMethods(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	resp, err := getByUser(user.ID, page, size)
	if err != nil {
		log.Errorf("User-UserTradingMethods-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": resp,
	})
}

func getByUser(userID uint64, page int32, size int32) (responseModel, error) {
	offset := (page - 1) * size

	var tradingMethod []*model.TradingMethod
	// 获取所有的交易方式
	if err := model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false).Find(&tradingMethod).Error; err != nil {
		log.Errorf("User-GetByUser-Error: %s", err.Error())
		return responseModel{}, errors.New("获取交易方式信息失败")
	}

	tradingMethodMap := make(map[uint64]*model.TradingMethod)
	if len(tradingMethod) > 0 {
		for _, method := range tradingMethod {
			tradingMethodMap[method.ID] = method
		}
	}

	tradingMethods := []*model.UserTradingMethodInfo{}

	baseQuery := model.DB.Table("user_trading_methods").Where(model.UserTradingMethod{UserID: userID, IsDeleted: false}).Order("created_at DESC")

	if err := baseQuery.Offset(offset).Limit(size).Find(&tradingMethods).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("User-GetByUser-Error: %s", err.Error())
		return responseModel{}, errors.New("获取用户支付方式列表失败")
	}

	// 获取总数量
	var totalCount int64
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("User-GetByUser-Error: %s", err.Error())
		return responseModel{}, errors.New("获取用户支付方式列表失败")
	}

	for _, method := range  tradingMethods {
		method.TradingMethodName = tradingMethodMap[method.TradingMethodID].Name
		method.TradingMethodEnName = tradingMethodMap[method.TradingMethodID].EnName
		method.TradingMethodIcon = tradingMethodMap[method.TradingMethodID].Icon
	}

	return responseModel {
		TradingMethods: tradingMethods,
		PageNo: page,
		PageSize: size,
		TotalPage: math.Ceil(float64(totalCount) / float64(size)),
		TotalCount: totalCount,
	}, nil
}

type responseModel struct {
	TradingMethods 		[]*model.UserTradingMethodInfo	`json:"tradingMethods"`
	PageNo				int32							`json:"pageNo"`
	PageSize 			int32							`json:"pageSize"`
	TotalPage 			float64							`json:"totalPage"`
	TotalCount 			int64							`json:"totalCount"`
}