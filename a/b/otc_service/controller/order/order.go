package order

import (
	"math"
	"time"
	"errors"
	"strconv"
	"strings"
	"math/big"
	"net/http"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/data"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"
	"github.com/spf13/viper"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"

	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
	taskApi "galaxyotc/common/service/task_service"
	"galaxyotc/gc_services/otc_service/api"
)

// 订单操作
type OrderOperation struct {
	Sn string `json:"sn"` // 订单流水号
}

// NewOrder 创建新订单
func NewOrder(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var order model.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	/*if !user.IsRealName {
		SendErrJSON("请先进行实名认证", c)
		return
	}*/

	var offer model.Offer
	// 查询广告来源是否有效
	if model.DB.Model(model.Offer{}).Where("code = ? AND status = ?", order.Source, model.OfferOnInt).First(&offer).RecordNotFound() {
		SendErrJSON("无效的广告", c)
		return
	}

	// 如果是出售广告，那么订单就是购买订单，如果是购买广告，订单就是出售订单
	if offer.OfferType == model.SellOfferInt {
		order.Seller = offer.AccountID
		order.OrderType = model.BuyOrderInt
		order.Buyer = user.ID
	} else if offer.OfferType == model.BuyOfferInt {
		order.Buyer = offer.AccountID
		order.OrderType = model.SellOrderInt
		order.Seller = user.ID
		// 出售订单需要提供出售者的支付方式便于买方付款
		if order.TradingMethod == 0 {
			SendErrJSON("无效的支付方式", c)
			return
		}
	}

	// 判断买方和卖方是否是同一用户
	if order.Buyer == order.Seller {
		SendErrJSON("不能自卖自买", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(offer.Currency)
	if err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 检查是否缺少必要的参数
	if order.Source == "" || order.TotalPrice == 0 {
		SendErrJSON("缺少必要的参数", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true) //payments.Fetcher.GetAllRates(true)
	if err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 获取当前的浮动汇率
	fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
	cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
	currentRate := fiatRate.DivRound(cryptoRate, 2)
	order.Price, _ = currentRate.Float64()

	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断
	if utils.IsNaNOrInf(order.Price) {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 单价等于当前汇率 * 逆价浮动
	price := currentRate.Mul(decimal.NewFromFloat(offer.Float))

	// 订单数量等于总金额 / (当前汇率 * 逆价浮动)
	amount := decimal.NewFromFloat(order.TotalPrice).DivRound(price, int32(currency.Precision))
	order.Amount, _ = amount.Float64()
	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
	if utils.IsNaNOrInf(order.Amount) {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("计算订单数量失败", c)
		return
	}

	// 校验金额是否在最小限额和最大限额之间
	if order.TotalPrice < offer.MinLimit || order.TotalPrice > offer.MaxLimit {
		SendErrJSON("交易金额不在限额范围内", c)
		return
	}

	var (
		buyer  model.User
		seller model.User
	)

	// 获取买家的个人信息
	if err := model.DB.Model(model.User{}).Where("id = ?", order.Buyer).First(&buyer).Error; err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("获取用户信息失败", c)
		return
	}

	// 获取卖家的个人信息
	if err := model.DB.Model(model.User{}).Where("id = ?", order.Seller).First(&seller).Error; err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("获取用户信息失败", c)
		return
	}

	// 将小数转换为位
	amountWei := utils.ToWei(order.Amount, int(currency.Precision))
	// 计算手续费 TODO 每个用户的手续费可能不一致
	gas := 0.11
	// 转成整数计算
	feeWei := float64(amountWei.Int64()) * gas
	// 转回指定精度的小数
	fee, _ := utils.ToDecimal(big.NewInt(int64(feeWei)), int(currency.Precision)).Float64()

	order.Currency = offer.Currency
	order.FiatCurrency = offer.FiatCurrency
	order.Fee = fee
	// 根据当前时间生成流水号
	now := time.Now().Local()
	// 添加订单流水号前缀
	order.Sn = viper.GetString("server.order_prefix") + xid.NewWithTime(now).String()  //Server.OrderPrefix + xid.NewWithTime(now).String()
	order.Status = model.OrderWaitingInt
	order.BuyerAddress = buyer.InternalAddress
	order.SellerAddress = seller.InternalAddress

	tx := model.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("创建订单失败", c)
		return
	}

	//创建订单状态明细
	var orderEvent model.OrderEvent
	orderEvent.OrderID = order.ID
	orderEvent.OrderType = order.OrderType
	orderEvent.Status = model.OrderWaitingInt
	orderEvent.InvalidTime =time.Now().Local().Add(time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute)
	orderEvent.TurnTime = time.Now().Local()
	if err := tx.Create(&orderEvent).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON("创建订单状态明细失败", c)
		return
	}

	tx.Commit()

	// 创建完订单后直接返回订单明细
	var orderInfo model.OrderInfo
	if err := orderInfo.Init(order.Price, &user, &offer, &order, &orderEvent); err != nil {
		log.Errorf("Order-NewOrder-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	// 推送通知给广告主
	go func() {
		userIDStr := strconv.FormatUint(uint64(offer.AccountID), 10)
		msgMap := utils.StructToMapInterface(orderInfo)
		// 推送通知给买方
		api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "您有新的订单", "请及时处理新订单", msgMap, 0)
		duration := time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute
		// 订单期限内卖家没接单就自动取消
		api.TaskApi.OrderTimeout(duration.Nanoseconds(), taskApi.OrderCancel, order.Sn, taskApi.WaitingApproved)
	}()

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"orderInfo": orderInfo,
		},
	})
}

// 根据筛选条件获取订单列表
func Orders(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取订单类型，默认为所有
	orderType := c.DefaultQuery("type", "all")

	// 获取订单状态,默认为交易中
	orderStatus := c.DefaultQuery("status", "all")

	// 获取代币类型
	//currencyStr := c.Query("currency")

	// 获取订单开始日期
	startTime := c.Query("start")

	// 获取订单结束日期
	endTime := c.Query("end")

	// 获取订单流水号
	orderSn := c.Query("sn")

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	//var currency model.Currency

	// 根据代币代码获取代币信息
	//if err := model.DB.Model(model.Currency{}).Where("code = ? and fiat = false and status = 0", strings.ToUpper(currencyStr)).First(&currency).Error; err != nil {
	//	fmt.Println(err.Error())
	//	SendErrJSON("获取代币信息失败", c)
	//	return
	//}

	var tradingMethod []*model.TradingMethod
	// 获取所有的交易方式
	if err := model.DB.Model(model.TradingMethod{}).Where("is_deleted = ?", false).Find(&tradingMethod).Error; err != nil {
		log.Errorf("Order-Orders-Error: %s", err.Error())
		SendErrJSON("获取交易方式信息失败", c)
		return
	}

	tradingMethodMap := make(map[uint64]*model.TradingMethod)
	if len(tradingMethod) > 0 {
		for _, method := range tradingMethod {
			tradingMethodMap[method.ID] = method
		}
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var (
		selectSQL string
		args      []interface{}
	)

	// 是否进行订单类型筛选
	switch orderType {
	case "all":
		selectSQL = "(seller = ? OR buyer = ?)"
		args = append(args, user.ID, user.ID)
	case "buy": // 购买订单
		selectSQL = "buyer = ?"
		args = append(args, user.ID)
	case "sell":
		selectSQL = "seller = ?"
		args = append(args, user.ID)
	default:
		SendErrJSON("无效的订单类型", c)
		return
	}

	// 是否进行订单状态筛选
	if orderStatus != "" {
		switch orderStatus {
		case "all":
			// 不作处理
		case "trading":
			selectSQL += " AND (status = ? OR status = ? OR status = ? OR status = ?)"
			args = append(args, model.OrderWaitingInt, model.OrderWaitingPayInt, model.OrderWaitingReleaseInt, model.OrderWaitingCompleteInt)
		case "completed":
			selectSQL += " AND status = ?"
			args = append(args, model.OrderCompletedInt)
		case "canceled":
			selectSQL += " AND status = ?"
			args = append(args, model.OrderCanceledInt)
		default:
			SendErrJSON("无效的订单状态", c)
			return
		}
	}

	// 是否进行订单日期筛选
	if startTime != "" && endTime == "" { // 开始日期不为空，结束日期为空，则筛选在开始日期之后的订单
		selectSQL += " AND done_at >= ?"
		args = append(args, startTime)
	} else if startTime == "" && endTime != "" { // 开始日期为空，结束日期不为空，则筛选在结束日期之前的订单
		selectSQL += " AND done_at <= ?"
		args = append(args, endTime)
	} else if startTime != "" && endTime != "" {
		selectSQL += " AND done_at BETWEEN ? and ?" // 开始日期不为空，结束日期不为空，则筛选在开始日期和结束日期之间的订单
		args = append(args, startTime, endTime)
	}

	// 是否指定了订单流水号
	if orderSn != "" {
		selectSQL += " AND sn = ?"
		args = append(args, orderSn)
	}

	var (
		orders     []*model.Order
		totalCount int64
	)

	baseQuery := model.DB.Model(&model.Order{}).Where(selectSQL, args...).Order("created_at DESC")

	if err := baseQuery.Offset(offset).Limit(size).Find(&orders).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Order-Orders-Error: %s", err.Error())
		SendErrJSON("获取订单列表失败", c)
		return
	}

	// 获取订单总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Order-Orders-Error: %s", err.Error())
		SendErrJSON("获取订单列表失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true) //payments.Fetcher.GetAllRates(true)
	if err != nil {
		log.Errorf("Order-Orders-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	orderList := []*model.OrderInfo{}

	if totalCount > 0 {
		for _, order := range orders {
			//根据广告码获取广告信息
			var offer model.Offer

			if err := model.DB.Model(&model.Offer{}).Where("code = ?", order.Source).First(&offer).Error; err == gorm.ErrRecordNotFound {
				SendErrJSON("广告不存在", c)
				return
			} else if err != nil {
				log.Errorf("Order-Orders-Error: %s", err.Error())
				SendErrJSON("获取广告详情失败", c)
				return
			}

			// 获取当前的浮动汇率
			fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
			cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
			currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

			// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
			if utils.IsNaNOrInf(currentPrice) {
				log.Errorf("Order-Orders-Error: %s", err.Error())
				SendErrJSON("获取代币汇率失败", c)
				return
			}

			var orderInfo model.OrderInfo
			if err := orderInfo.Init(currentPrice, &user, &offer, order, nil); err != nil {
				log.Errorf("Order-Orders-Error: %s", err.Error())
				SendErrJSON(err.Error(), c)
				return
			}

			orderList = append(orderList, &orderInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"orders":     orderList,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

// 查看订单明细
func OrderDetail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	sn := c.Param("sn")
	if sn == "" {
		SendErrJSON("订单流水号不能为空", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var order model.Order

	if err := model.DB.Where("sn = ?", sn).First(&order).Error; err != nil {
		log.Errorf("Order-OrderDetail-Error: %s", err.Error())
		SendErrJSON("获取订单明细失败", c)
		return
	}

	//根据广告码获取广告信息
	var offer model.Offer

	if err := model.DB.Model(&model.Offer{}).Where("code = ?", order.Source).First(&offer).Error; err == gorm.ErrRecordNotFound {
		SendErrJSON("广告不存在", c)
		return
	} else if err != nil {
		log.Errorf("Order-OrderDetail-Error: %s", err.Error())
		SendErrJSON("获取广告详情失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Order-OrderDetail-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
	}

	// 获取当前的浮动汇率
	fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
	cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
	currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
	if utils.IsNaNOrInf(currentPrice) {
		log.Errorf("Order-OrderDetail-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
	}

	var orderInfo model.OrderInfo
	if err := orderInfo.Init(currentPrice, &user, &offer, &order, nil); err != nil {
		log.Errorf("Order-OrderDetail-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"orderInfo": orderInfo,
		},
	})
}

// 获取最近完成的交易记录
func RecentlyOrders(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取代币类型
	currencyCode := strings.ToUpper(c.DefaultQuery("currency", "btc"))

	// 获取法币类型
	fiatCurrencyCode := strings.ToUpper(c.DefaultQuery("fiat_currency", "cny"))

	// 获取返回数量,默认十条
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		log.Errorf("Order-RecentlyOrders-Error: %s", err.Error())
		SendErrJSON("无效的参数", c)
	}

	var orders []*model.Order

	if err := model.DB.Table("orders").Where("currency = ? AND fiat_currency = ? AND status = ?", currencyCode, fiatCurrencyCode, model.OrderCompletedInt).Order("done_at DESC").Limit(limit).Find(&orders).Error; err != nil {
		log.Errorf("Order-RecentlyOrders-Error: %s", err.Error())
		SendErrJSON("获取最近交易列表失败", c)
		return
	}

	recentlyOrders := []*model.RecentlyOrder{}

	if len(orders) > 0 {
		for _, order := range orders {
			recentlyOrders = append(recentlyOrders, &model.RecentlyOrder{
				Currency:         order.Currency,
				FiatCurrency:     order.FiatCurrency,
				Amount:           order.Amount,
				TotalPrice:       order.TotalPrice,
				OrderType:        order.OrderType,
				DoneAt:           *order.DoneAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"recentlyOrders": recentlyOrders,
		},
	})
}

// 推送订单明细
func PushOrderDetail(user model.User, order model.Order) (msgMap map[string]interface{}, err error) {
	//根据广告码获取广告信息
	var offer model.Offer

	if err := model.DB.Model(&model.Offer{}).Where("code = ?", order.Source).First(&offer).Error; err != nil {
		return nil, err
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		return nil, err
	}

	// 获取当前的浮动汇率
	fiatRate := decimal.NewFromFloat(allRates[offer.FiatCurrency])
	cryptoRate := decimal.NewFromFloat(allRates[offer.Currency])
	currentPrice, _ := fiatRate.DivRound(cryptoRate, 2).Float64()

	// 当断网或获取不到汇率时，值会变得异常，所以需要进行判断处理
	if utils.IsNaNOrInf(currentPrice) {
		return nil, errors.New("汇率计算异常")
	}

	var orderInfo model.OrderInfo
	if err := orderInfo.Init(currentPrice, &user, &offer, &order, nil); err != nil {
		return nil, err
	}

	msgMap = utils.StructToMapInterface(orderInfo)
	return msgMap, nil
}

// 卖家或买家接单，冻结代币
func Approved(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var orderOperation OrderOperation
	if err := c.ShouldBindJSON(&orderOperation); err != nil {
		log.Errorf("Order-Approved-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order
	// 获取订单信息
	if err := model.DB.Where("sn = ?", orderOperation.Sn).First(&order).Error; err != nil {
		log.Errorf("Order-Approved-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 确认订单前的订单状态应该是等待接单
	if order.Status != model.OrderWaitingInt {
		SendErrJSON("订单状态无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var (
		traderID  uint64
		pushTitle string
	)
	// 判断订单类型
	if order.OrderType == model.BuyOrderInt {
		// 卖家同意订单，进行冻结代币
		if order.Seller != user.ID {
			SendErrJSON("您没有权限执行此操作", c)
			return
		}
		pushTitle = "对方已接单，请付款"
		traderID = order.Buyer
	} else if order.OrderType == model.SellOrderInt {
		// 买家同意订单，进行冻结代币
		if order.Buyer != user.ID {
			SendErrJSON("您没有权限执行此操作", c)
			return
		}
		pushTitle = "对方已接单，请等待"
		traderID = order.Seller
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(order.Currency)
	if err != nil {
		log.Errorf("Order-Approved-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
		return
	}

	// 将小数转换为位
	amountWei := utils.ToWei(order.Amount, int(currency.Precision))
	// 将订单流水号的前缀去掉
	sn := strings.TrimLeft(order.Sn, viper.GetString("otc.order_prefix"))
	txid, err := api.PrivateWalletApi.AddTokenTransaction(amountWei.String(), sn, 2, 0, order.BuyerAddress, order.SellerAddress, currency.PrivateTokenAddress)
	if err != nil {
		log.Errorf("Order-Approved-Error: %s", err.Error())
		if strings.Contains(err.Error(), "余额不足") {
			SendErrJSON("确认订单失败，余额不足", c)
			return
		} else {
			SendErrJSON("锁币失败", c)
			return
		}
	}

	tx := model.DB.Begin()

	// 接单后更改状态为等待付款
	if err := tx.Model(&order).Update("status", model.OrderWaitingPayInt).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-Approved-Error: %s", err.Error())
		SendErrJSON("保存锁币交易数据失败", c)
		return
	}

	// 创建订单状态明细
	var orderEvent model.OrderEvent
	orderEvent.OrderID = order.ID
	orderEvent.OrderType = order.OrderType
	orderEvent.Status = model.OrderWaitingPayInt
	orderEvent.InvalidTime = time.Now().Local().Add(time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute)
	orderEvent.TurnTime = time.Now().Local()
	orderEvent.Txid = txid
	if err :=tx.Create(&orderEvent).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-Approved-Error: %s", err.Error())
		SendErrJSON("创建订单状态明细失败", c)
		return
	}

	tx.Commit()

	// 推送通知给对方交易人
	go func() {
		userIDStr := strconv.FormatUint(uint64(traderID), 10)
		msgMap, _ := PushOrderDetail(user, order)
		//推送通知给买方
		api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, pushTitle, "点击查看详情", msgMap, 0)
		duration := time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute
		// 订单期限内买家没付款就自动取消
		api.TaskApi.OrderTimeout(duration.Nanoseconds(), taskApi.OrderCancel, order.Sn, taskApi.WaitingPay)
	}()

	// 计算订单倒计时
	t1 := time.Now().Local()
	t2 := orderEvent.InvalidTime

	subM := t2.Sub(t1)
	countDownSeconds := int(subM.Seconds())

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{
			"seconds": countDownSeconds,
		},
	})
}

// 买家确认付款，修改订单状态
func CompletedPay(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var orderOperation OrderOperation
	if err := c.ShouldBindJSON(&orderOperation); err != nil {
		log.Errorf("Order-CompletedPay-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order

	// 获取订单信息
	if err := model.DB.Where("sn = ?", orderOperation.Sn).First(&order).Error; err != nil {
		log.Errorf("Order-CompletedPay-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 确认付款前的订单状态应该是等待付款
	if order.Status != model.OrderWaitingPayInt {
		SendErrJSON("订单状态无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 判断当前用户是否是买家
	if order.Buyer != user.ID {
		SendErrJSON("您没有权限执行此操作", c)
		return
	}

	tx := model.DB.Begin()

	// 确认付款后更改状态为等待放币
	if err := tx.Model(&order).Update("status", model.OrderWaitingReleaseInt).Error; err != nil {
		log.Errorf("Order-CompletedPay-Error: %s", err.Error())
		SendErrJSON("保存交易状态数据失败", c)
		return
	}

	// 创建订单状态明细
	var orderEvent model.OrderEvent
	orderEvent.OrderID = order.ID
	orderEvent.OrderType = order.OrderType
	orderEvent.Status = model.OrderWaitingReleaseInt
	orderEvent.InvalidTime = time.Now().Local().Add(time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute)
	orderEvent.TurnTime = time.Now().Local()
	if err := tx.Create(&orderEvent).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-CompletedPay-Error: %s", err.Error())
		SendErrJSON("创建订单状态明细失败", c)
		return
	}

	tx.Commit()

	// 推送通知给卖方
	go func() {
		userIDStr := strconv.FormatUint(uint64(order.Seller), 10)
		msgMap, _ := PushOrderDetail(user, order)
		//推送通知给买方
		api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "买方已付款", "请确认收到款后放币", msgMap, 0)
		duration := time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute
		// 订单期限内卖家没放币就自动放币
		api.TaskApi.OrderTimeout(duration.Nanoseconds(), taskApi.OrderRelease, order.Sn)
	}()

	// 计算订单倒计时
	t1 := time.Now().Local()
	t2 := orderEvent.InvalidTime

	subM := t2.Sub(t1)
	countDownSeconds := int(subM.Seconds())

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{
			"seconds": countDownSeconds,
		},
	})
}

func executeTransaction(amountWei, orderSn, buyerAddress, sellerAddress, tokenAddress string, timeout int64, toUser int32) (string, error) {
	// 将订单流水号的前缀去掉
	sn := strings.TrimLeft(orderSn, viper.GetString("otc.order_prefix"))

	txid, err := api.PrivateWalletApi.ExecuteTransaction(amountWei, sn, 2, timeout, buyerAddress, sellerAddress, tokenAddress, toUser)
	if err != nil {
		return "", err
	}

	return txid, nil
}

// 同意放币，转给买家
func Release(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var orderOperation OrderOperation
	if err := c.ShouldBindJSON(&orderOperation); err != nil {
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order

	// 获取订单信息
	if err := model.DB.Where("sn = ?", orderOperation.Sn).First(&order).Error; err != nil {
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 确认放币前的订单状态应该是等待放币
	if order.Status != model.OrderWaitingReleaseInt {
		SendErrJSON("订单状态无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 只有卖家同意订单才进行释放代币
	if order.Seller != user.ID {
		SendErrJSON("您没有权限执行此操作", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(order.Currency)
	if err != nil {
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
		return
	}

	// 获取卖家的余额并判断是否足够
	balanceWei, err := api.PrivateWalletApi.GetTokenBalance(currency.PrivateTokenAddress, user.InternalAddress)
	if err != nil {
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("获取卖家余额失败", c)
		return
	}

	// 将余额由位转成小数
	blance, _ := utils.ToDecimal(balanceWei, int(currency.Precision)).Float64()

	if blance < order.Amount {
		// 余额不够提示用户充值
		go func() {
			userIDStr := strconv.FormatUint(uint64(user.ID), 10)
			msgMap, _ := PushOrderDetail(user, order)
			//推送通知给买方
			api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "余额不足，无法进行交易", "请及时充值", msgMap, 0)
		}()
		SendErrJSON("账户余额不足，请先进行充值", c)
		return
	}

	// 将小数转换为位
	amountWei := utils.ToWei(order.Amount, int(currency.Precision))
	// 执行多签名放币操作
	txid, err := executeTransaction(amountWei.String(), order.Sn, order.BuyerAddress, order.SellerAddress, currency.PrivateTokenAddress, 0, privateWalletApi.ToBuyer)
	if err != nil {
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("交易放币失败,请联系客服", c)
		return
	}

	tx := model.DB.Begin()
	if err := tx.Model(&order).Updates(model.Order{Txid: txid, Status: model.OrderWaitingCompleteInt}).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("保存订单交易数据失败", c)
		return
	}

	//创建订单状态明细
	var orderEvent model.OrderEvent
	orderEvent.OrderID = order.ID
	orderEvent.OrderType = order.OrderType
	orderEvent.Status = model.OrderWaitingCompleteInt
	orderEvent.InvalidTime = time.Now().Local().Add(time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute)
	orderEvent.TurnTime = time.Now().Local()
	orderEvent.Txid = txid
	if err :=tx.Create(&orderEvent).Error; err != nil {
		tx.Rollback()
		log.Errorf("Order-Release-Error: %s", err.Error())
		SendErrJSON("创建订单状态明细失败", c)
		return
	}

	tx.Commit()

	// 推送通知给买方以及订单自动确认倒计时
	go func() {
		userIDStr := strconv.FormatUint(uint64(order.Buyer), 10)
		msgMap, _ := PushOrderDetail(user, order)
		//推送通知给买方
		api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "卖方已放币", "请确认", msgMap, 0)
		duration := time.Duration(viper.GetInt("otc.order_max_age")) * time.Minute
		// 订单期限内买家没确认就自动确认
		api.TaskApi.OrderTimeout(duration.Nanoseconds(), taskApi.OrderCompleted, order.Sn)
	}()

	t1 := time.Now().Local()
	t2 := orderEvent.InvalidTime

	var countDownSeconds int
	if t1.Before(t2) {
		subM := t2.Sub(t1)
		countDownSeconds = int(subM.Seconds())
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{
			"seconds": countDownSeconds,
		},
	})
}

// 取消放币，退回卖家
func Canceled(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var orderOperation OrderOperation
	if err := c.ShouldBindJSON(&orderOperation); err != nil {
		log.Errorf("Order-Canceled-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order

	// 获取订单信息
	if err := model.DB.Where("sn = ?", orderOperation.Sn).First(&order).Error; err != nil {
		log.Errorf("Order-Canceled-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	if order.Status == model.OrderWaitingInt { // 等待接单中,广告者取消订单
		// 根据订单类型,判断用户是否是广告发布者
		var traderID uint64
		if order.OrderType == model.BuyOrderInt {
			// 购买订单，卖方是广告者
			if order.Seller != user.ID {
				SendErrJSON("您没有权限执行此操作", c)
				return
			}
			traderID = order.Buyer
		} else if order.OrderType == model.SellOrderInt {
			// 出售订单，买方是广告者
			if order.Buyer != user.ID {
				SendErrJSON("您没有权限执行此操作", c)
				return
			}
			traderID = order.Seller
		}

		doneAt := time.Now().Local()
		if err := model.DB.Model(&order).Updates(model.Order{Status: model.OrderCanceledInt, CancelBy: user.ID, DoneAt: &doneAt}).Error; err != nil {
			log.Errorf("Order-Canceled-Error: %s", err.Error())
			SendErrJSON("取消订单失败", c)
			return
		}

		// 推送通知给对方交易人
		go func() {
			userIDStr := strconv.FormatUint(uint64(traderID), 10)
			msgMap, _ := PushOrderDetail(user, order)
			//推送通知给对方
			api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "交易取消", "对方取消了订单", msgMap, 0)
		}()

	} else if order.Status == model.OrderWaitingPayInt {
		if order.Buyer == user.ID {
			// 获取代币信息
			currency, err := model.CurrencyFromAndToRedis(order.Currency)
			if err != nil {
				log.Errorf("Order-Canceled-Error: %s", err.Error())
				SendErrJSON("获取代币信息失败", c)
				return
			}

			// 将小数转换为位
			amountWei := utils.ToWei(order.Amount, int(currency.Precision))
			txid, err := executeTransaction(amountWei.String(), order.Sn, order.BuyerAddress, order.SellerAddress, currency.PrivateTokenAddress, 0, privateWalletApi.ToSeller)
			if err != nil {
				log.Errorf("Order-Canceled-Error: %s", err.Error())
				SendErrJSON("交易放币失败,请联系客服", c)
				return
			}

			doneAt := time.Now().Local()
			if err := model.DB.Model(&order).Updates(model.Order{Txid: txid, Status: model.OrderCanceledInt, CancelBy: user.ID, DoneAt: &doneAt}).Error; err != nil {
				log.Errorf("Order-Canceled-Error: %s", err.Error())
				SendErrJSON("取消订单失败", c)
				return
			}

			// 推送通知给卖方
			go func() {
				userIDStr := strconv.FormatUint(uint64(order.Seller), 10)
				msgMap, _ := PushOrderDetail(user, order)
				//推送通知给对方
				api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "交易取消", "买方取消了付款", msgMap, 0)
			}()
		} else {
			SendErrJSON("您没有权限执行此操作", c)
			return
		}
	} else {
		SendErrJSON("无效的订单状态", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 买家确认收币，订单已完成
func Completed(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var orderOperation OrderOperation
	if err := c.ShouldBindJSON(&orderOperation); err != nil {
		log.Errorf("Order-Completed-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order
	// 获取订单信息
	if err := model.DB.Where("sn = ?", orderOperation.Sn).First(&order).Error; err != nil {
		log.Errorf("Order-Completed-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 确认收币前的订单状态应该是等待确认
	if order.Status != model.OrderWaitingCompleteInt {
		SendErrJSON("订单状态无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 买家确认订单
	if order.Buyer != user.ID {
		SendErrJSON("您没有权限执行此操作", c)
		return
	}

	// 确认后更改状态为交易完成
	doneAt := time.Now().Local()
	if err := model.DB.Model(&order).Updates(model.Order{Status: model.OrderCompletedInt, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("Order-Completed-Error: %s", err.Error())
		SendErrJSON("修改订单状态失败", c)
		return
	}

	// 推送通知给卖方
	go func() {
		userIDStr := strconv.FormatUint(uint64(order.Seller), 10)
		msgMap, _ := PushOrderDetail(user, order)
		//推送通知给买方
		api.PushApi.SendMsg("notification", userIDStr, data.APPID_OTC, "买方已确认收币", "交易完成", msgMap, 0)

		// TODO: 添加一个字段保存每单利润
		// 分佣计算
		//var offerUserID uint
		//if order.OrderType == model.BuyOrderInt {
		//	offerUserID = order.Seller
		//} else  if order.OrderType == model.SellOrderInt {
		//	offerUserID = order.Buyer
		//}
		//commission_distribution.ApplyByOrder(offerUserID, order.ID, order.Currency, order.Profit)
	}()

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

type arbitrationReq struct {
	OrderCode string 	`json:"order_code"`
	WinnerID  uint64   	`json:"winner_id"`
}

// 仲裁放币，平均分配
func Arbitration(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req arbitrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("Order-Arbitration-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	var order model.Order

	// 获取订单信息
	if err := model.DB.Where("code = ?", req.OrderCode).First(&order).Error; err != nil {
		log.Errorf("Order-Arbitration-Error: %s", err.Error())
		SendErrJSON("无效的订单", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(order.Currency)
	if err != nil {
		log.Errorf("Order-Arbitration-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
		return
	}

	var toUser int32
	switch req.WinnerID {
	case order.Buyer:
		toUser = privateWalletApi.ToBuyer
	case order.Seller:
		toUser = privateWalletApi.ToSeller
	default:
		SendErrJSON("无效的胜诉人", c)
		return
	}

	// 将小数转换为位
	amountWei := utils.ToWei(order.Amount, int(currency.Precision))
	txid, err := executeTransaction(amountWei.String(), order.Sn, order.BuyerAddress, order.SellerAddress, currency.PrivateTokenAddress, 0, toUser)
	if err != nil {
		log.Errorf("Order-Arbitration-Error: %s", err.Error())
		SendErrJSON("交易放币失败", c)
		return
	}

	//TODO 修改仲裁表
	doneAt := time.Now().Local()
	// 将订单状态修改为已仲裁
	if err := model.DB.Model(&order).Updates(model.Order{Txid: txid, Status: model.OrderArbitrationInt, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("Order-Arbitration-Error: %s", err.Error())
		SendErrJSON("修改订单状态失败", c)
		return
	}

	// TODO 仲裁成功后推送消息给双方
	go func() {
	}()

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}
