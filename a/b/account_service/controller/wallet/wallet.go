package wallet

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/net"
	"galaxyotc/common/utils"

	"galaxyotc/gc_services/account_service/api"
	"github.com/rs/xid"
)

// 资产管理-钱包信息
type Wallet struct {
	//Currency        uint64  `json:"currency"`         // 代币ID
	Balance         float64 `json:"balance"`          // 流动余额
	Freezed         float64 `json:"freezed"`          // 冻结金额
	TotalAmount     float64 `json:"total_amount"`     // 总额
	AssetsValuation float64 `json:"assets_valuation"` // 资产估值人民币
	AssetsBTC       float64 `json:"assets_btc"`       // 资产估值比特币
	Name            string  `json:"name"`             // 币种名
	Code            string  `json:"code"`             // 币种代码
	Icon            string  `json:"icon"`             // 图标
	Sort            int32   `json:"-"`
}

func (wallet *Wallet) Init(user *model.User, currency *model.Currency, allRates map[string]float64) error {
	wallet.Code = currency.Code
	wallet.Name = currency.Name
	wallet.Icon = currency.Icon
	wallet.Sort = currency.Sort

	// 获取对应代币的余额
	blanceWeiStr, err := api.PrivateWalletApi.GetTokenBalance(currency.PrivateTokenAddress, user.InternalAddress)
	if err != nil {
		log.Errorf("Wallets Error: %s", err.Error())
		return err
	}

	// 根据卖方指定币种的待确认的订单状态获取被冻结的余额
	var amountList []float64
	if err := model.DB.Model(&model.Order{}).Where("currency = ? AND seller = ? AND status = ?", currency.Code, user.ID, model.OrderWaitingCompleteInt).Pluck("amount", &amountList).Error; err != nil {
		log.Errorf("Wallets Error: %s", err.Error())
		return errors.New("获取冻结余额失败")
	}

	// 为控制精度将小数转成整数的位在进行计算
	freezedWei := new(big.Int)
	for _, amount := range amountList {
		amountWei := utils.ToWei(amount, int(currency.Precision))
		// 累加冻结余额
		freezedWei.Add(freezedWei, amountWei)
	}

	blanceWei, _ := new(big.Int).SetString(blanceWeiStr, 10)

	var places int32
	// ETH和BTC保留8位小数，EOS保留4位
	switch currency.Family {
	case "BTC":
		places = 8
	case "ETH":
		places = 8
	case "EOS":
		places = 4
	}

	// 可用余额不需要计算
	wallet.Balance, _ = utils.ToDecimal(blanceWei, int(currency.Precision)).Round(places).Float64()

	// 将计算后的位转回小数
	wallet.Freezed, _ = utils.ToDecimal(freezedWei, int(currency.Precision)).Round(places).Float64()

	// 总额等于可用余额加上冻结余额
	wallet.TotalAmount, _ = utils.ToDecimal(blanceWei.Add(blanceWei, freezedWei), int(currency.Precision)).Round(places).Float64()

	if wallet.TotalAmount != 0 {
		// 单个币种对应人民币的价值
		fiatRate := decimal.NewFromFloat(allRates["CNY"])
		cryptoRate := decimal.NewFromFloat(allRates[wallet.Code])
		currencyValuation := fiatRate.DivRound(cryptoRate, 2)

		totalAmountDecimal := decimal.NewFromFloat(wallet.TotalAmount)

		// 资产估值人民币 = 该币种总资产 * 单个价值
		wallet.AssetsValuation, _ = totalAmountDecimal.Mul(currencyValuation).Round(2).Float64()

		// 资产估值比特币 = 该币种总资产 / 一个比特币对应的价值
		wallet.AssetsBTC, _ = totalAmountDecimal.DivRound(cryptoRate, 8).Float64()

		// 当断网或获取不到汇率时，资产估值会变成异常，所以需要进行判断处理
		if utils.IsNaNOrInf(wallet.AssetsValuation) || utils.IsNaNOrInf(wallet.AssetsBTC) {
			return errors.New("计算资产估值失败")
		}
	}
	return nil
}

// 钱包列表, 实现sort包排序方法
type WalletList []*Wallet

// 获取此 slice 的长度
func (w WalletList) Len() int {
	return len(w)
}

// 根据钱包的资产估值降序排序
//func (w WalletList) Less(i, j int) bool {
//	if w[i].AssetsValuation == w[j].AssetsValuation {
//		return w[i].TotalAmount > w[j].TotalAmount
//	}
//	return w[i].AssetsValuation > w[j].AssetsValuation
//}

// 根据币种排序
func (w WalletList) Less(i, j int) bool {
	return w[i].Sort < w[j].Sort
}

// 交换数据
func (w WalletList) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

// Wallets 获取账户钱包列表
func Wallets(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var currencies []*model.Currency

	// 不能使用var声明不然为空时返回null而不是空列表
	wallets := WalletList{}

	// 获取所有币种信息
	if err := model.DB.Model(model.Currency{}).Where("status = 0").Find(&currencies).Error; err != nil {
		log.Errorf("Account-Wallets-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Account-Wallets-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 用户总资产估值
	var (
		totalAssetsValuation float64
		totalAssetsBTC       float64
	)

	// 循环代币，根据每一个代币的token_address和用户的内部私有地址获取用户每一个币种的余额和冻结金额
	for _, currency := range currencies {
		var wallet Wallet
		if err := wallet.Init(&user, currency, allRates); err != nil {
			log.Errorf("Account-Wallets-Error: %s", err.Error())
			SendErrJSON(err.Error(), c)
			return
		}

		// 计算人民币总价值和比特币总价值并将控制精度
		totalAssetsValuationDecimal := decimal.NewFromFloat(totalAssetsValuation).Add(decimal.NewFromFloat(wallet.AssetsValuation))
		totalAssetsValuation, _ = totalAssetsValuationDecimal.Round(2).Float64()

		totalAssetsBTCDecimal := decimal.NewFromFloat(totalAssetsBTC).Add(decimal.NewFromFloat(wallet.AssetsBTC))
		totalAssetsBTC, _ = totalAssetsBTCDecimal.Round(8).Float64()

		wallets = append(wallets, &wallet)
	}

	// 对钱包列表按资产估值进行排序
	sort.Sort(wallets)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"total_assets_valuation": totalAssetsValuation,
			"total_assets_btc":       totalAssetsBTC,
			"wallets":                wallets,
		},
	})
}

// 获取指定币种的钱包信息
func CurrencyWallet(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	code := c.Query("code")
	if code == "" {
		SendErrJSON("代币代码不能为空", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(code)
	if err != nil {
		log.Errorf("Account-CurrencyWallet-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Account-CurrencyWallet-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var wallet Wallet
	if err := wallet.Init(&user, &currency, allRates); err != nil {
		log.Errorf("Account-CurrencyWallet-Error: %s", err.Error())
		SendErrJSON(err.Error(), c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"wallet": wallet,
		},
	})
}

type orderStatistics struct {
	Amount float64 `json:"amount"`
	Code   string  `json:"code"`
}

type periodStatistics struct {
	Bought []*orderStatistics `json:"bought"`
	Sold   []*orderStatistics `json:"sold"`
}

// 本周交易和历史交易计算逻辑相同，统一封装起来
func (p *periodStatistics) Statistics(userID uint64, orderList []*model.Order, currency *model.Currency) {
	if len(orderList) > 0 {
		// 为了控制精度将小数转成整数在进行计算
		var boughtAmountWei, soldAmountWei int64

		// 遍历订单列表，计算本周买入数量和卖出数量
		for _, order := range orderList {
			if order.Buyer == userID {
				boughtAmountWei += utils.ToWei(order.Amount, int(currency.Precision)).Int64()
			} else if order.Seller == userID {
				soldAmountWei += utils.ToWei(order.Amount, int(currency.Precision)).Int64()
			}
		}

		// 将计算后的整数转回小数
		boughtAmount, _ := utils.ToDecimal(big.NewInt(boughtAmountWei), int(currency.Precision)).Float64()
		soldAmount, _ := utils.ToDecimal(big.NewInt(soldAmountWei), int(currency.Precision)).Float64()

		// 添加到本周买入交易列表
		if boughtAmount != 0 {
			p.Bought = append(p.Bought, &orderStatistics{boughtAmount, currency.Code})
		}
		// 添加到本周卖出交易列表
		if soldAmount != 0 {
			p.Sold = append(p.Sold, &orderStatistics{soldAmount, currency.Code})
		}
	}
}

// 数据统计
func Statistics(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var currencies []*model.Currency

	// 获取所有币种信息
	if err := model.DB.Model(model.Currency{}).Where("status = 0").Find(&currencies).Error; err != nil {
		log.Errorf("Account-Statistics-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
		return
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Account-Statistics-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 本周交易,买入和卖出
	weekOrders := periodStatistics{
		Bought: []*orderStatistics{},
		Sold:   []*orderStatistics{},
	}

	// 历史交易,买入和卖出
	historyOrders := periodStatistics{
		Bought: []*orderStatistics{},
		Sold:   []*orderStatistics{},
	}

	// 资产总估值
	var (
		totalAssetsValuation float64
		totalAssetsBTC       float64
	)

	// 遍历代币列表，根据每一个代币的token_address和用户的内部私有地址获取用户每一个币种的余额，然后计算用户总资产, 并根据每一个币种获取对应的订单列表，计算交易总量
	for _, currency := range currencies {

		// 获取对应代币的余额
		balanceInt, err := api.PrivateWalletApi.GetTokenBalance(currency.PrivateTokenAddress, user.InternalAddress)
		if err != nil {
			log.Errorf("Account-Statistics-Error: %s", err.Error())
			SendErrJSON("获取代币余额失败", c)
			return
		}

		// 将余额由位转成小数
		balance, _ := utils.ToDecimal(balanceInt, int(currency.Precision)).Float64()

		if balance > 0 {
			// 单个币种对应人民币的价值
			fiatRate := decimal.NewFromFloat(allRates["CNY"])
			cryptoRate := decimal.NewFromFloat(allRates[currency.Code])
			currencyValuation := fiatRate.DivRound(cryptoRate, 2)

			fmt.Println(currency.Code)
			fmt.Println(fiatRate.String())
			fmt.Println(cryptoRate.String())
			fmt.Println(currencyValuation.String())

			balanceDecimal := decimal.NewFromFloat(balance)

			// 资产估值人民币 = 该币种总资产 * 单个价值
			assetsValuation, _ := balanceDecimal.Mul(currencyValuation).Round(2).Float64()
			// 资产估值比特币 = 该币种总资产 / 一个比特币对应的价值
			assetsBTC, _ := balanceDecimal.DivRound(cryptoRate, 8).Float64()

			// 计算人民币总价值和比特币总价值并将控制精度
			totalAssetsValuationDecimal := decimal.NewFromFloat(totalAssetsValuation).Add(decimal.NewFromFloat(assetsValuation))
			totalAssetsValuation, _ = totalAssetsValuationDecimal.Round(2).Float64()

			totalAssetsBTCDecimal := decimal.NewFromFloat(totalAssetsBTC).Add(decimal.NewFromFloat(assetsBTC))
			totalAssetsBTC, _ = totalAssetsBTCDecimal.Round(8).Float64()
		}

		// 本周交易订单列表
		var weekOrderList []*model.Order
		// 根据当前时间获取本周的第一天和最后一天
		firstDay, lastDay := utils.GetFirstDayAndLastDayOfWeek(time.Now().ISOWeek())
		// 本周订单
		if err := model.DB.Table("orders").Where("currency = ? AND status = ? AND (seller = ? OR buyer = ?) AND done_at BETWEEN ? AND ?", currency.Code, model.OrderCompletedInt, user.ID, user.ID, firstDay, lastDay).Find(&weekOrderList).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("Account-Statistics-Error: %s", err.Error())
			SendErrJSON("获取本周订单失败", c)
			return
		}

		// 本周交易统计
		weekOrders.Statistics(user.ID, weekOrderList, currency)

		// 历史交易订单列表
		var historyOrderList []*model.Order
		// 历史订单
		if err := model.DB.Table("orders").Where("currency =? AND status = ? AND (seller = ? OR buyer = ?)", currency.Code, model.OrderCompletedInt, user.ID, user.ID).Find(&historyOrderList).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("Account-Statistics-Error: %s", err.Error())
			SendErrJSON("获取历史订单失败", c)
			return
		}

		// 历史交易统计
		historyOrders.Statistics(user.ID, historyOrderList, currency)
	}

	// 当断网或获取不到汇率时，资产总估值会变成异常，所以需要进行判断处理
	if utils.IsNaNOrInf(totalAssetsValuation) || utils.IsNaNOrInf(totalAssetsBTC) {
		SendErrJSON("计算资产估值失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"total_assets_valuation": totalAssetsValuation,
			"total_assets_btc":       totalAssetsBTC,
			"week_orders":            weekOrders,
			"history_orders":         historyOrders,
		},
	})
}

var (
	depositType  int32 = 0
	withdrawType int32 = 1
)

// 充值/提币记录
type AccountHistories struct {
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	Sn           string    `json:"sn"`            // 交易流水号
	Currency     string    `json:"currency"`      // 代币代码
	Amount       float64   `json:"amount"`        // 代币数量
	Gas          float64   `json:"gas"`           // 矿工费
	Address      string    `json:"address"`       // 充值/提币地址
	Status       int32     `json:"status"`        // 状态ID
	StatusString string    `json:"status_string"` // 状态说明
	DoneAt       time.Time `json:"done_at"`       // 完成时间
	Txid         string    `json:"txid"`          // 交易哈希
	Type         int32     `json:"type"`          // 记录类型
}

// 根据筛选条件获取充值提币记录
func Histories(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 根据充值还是提币获取对应历史记录，默认为所有
	historyType := c.DefaultQuery("type", "all")

	// 获取代币类型
	code := c.DefaultQuery("code", "all")

	// 获取订单流水号
	Sn := c.Query("sn")

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)
	userID := user.ID

	var (
		selectSQL string
		args      []interface{}
	)

	// 过滤状态为未充值的充值记录
	args = append(args, model.DepositNotInt)
	// 是否进行订单类型筛选
	switch historyType {
	case "all":
		selectSQL = "histories.account_id = ?"
		args = append(args, userID)
	case "deposit": // 充值记录
		selectSQL = "histories.type = ? AND histories.account_id = ?"
		args = append(args, depositType, userID)
	case "withdraw": // 提币记录
		selectSQL = "histories.type = ? AND histories.account_id = ?"
		args = append(args, withdrawType, userID)
	default:
		SendErrJSON("无效的记录类型", c)
		return
	}

	// 是否指定了订单流水号
	if Sn != "" {
		selectSQL += " AND histories.sn = ?"
		args = append(args, Sn)
	}

	// 是否指定了代币代码
	if code != "all" {
		// 根据代币代码查询代币信息
		currency, err := model.CurrencyFromAndToRedis(strings.ToUpper(code))
		if err != nil {
			SendErrJSON("无效的代币类型", c)
			return
		}

		selectSQL += " AND histories.currency = ?"
		args = append(args, currency.Code)
	}

	var (
		histories  []*AccountHistories
		totalCount int64
	)

	//资产总估值
	var (
		outTotalAssetsBTC float64 // 提币总数
		inTotalAssetsBTC  float64 // 充值总数
	)

	var currencies []*model.Currency

	var (
		selectSQLCurrency string
		argsCurrency      []interface{}
	)
	selectSQLCurrency += "1=1"
	if code != "all" {
		selectSQLCurrency += " and (code = ?)"
		argsCurrency = append(args, code)
	}

	// 获取所有币种的最新汇率
	allRates, err := api.ExchangerateApi.GetAllRates(true)
	if err != nil {
		log.Errorf("Account-Histories-Error: %s", err.Error())
		SendErrJSON("获取代币汇率失败", c)
		return
	}
	// 获取所有币种信息
	if err := model.DB.Model(model.Currency{}).Where("status = 0").Where(selectSQLCurrency, argsCurrency...).Find(&currencies).Error; err != nil {
		log.Errorf("Account-Histories-Error: %s", err.Error())
		SendErrJSON("获取代币信息失败", c)
	}

	//var depositInfo []*model.DepositInfo
	//baseQuery := model.DB.Exec("SELECT SUM(amount)as totalAmount,MIN(currency) as bitcoin FROM withdraws where account_id=? GROUP BY currency", userID).Select("dep.*")

	//if err := baseQuery.Find(&depositInfo).Error; err != nil {
	//	log.Errorf("Account-Histories-Error: %s", err.Error())
	//}
	type WithdrawsInfo struct {
		Amount float64 `json:"amount"`
	}

	for _, currencieItem := range currencies {

		//fiatRate := decimal.NewFromFloat(allRates["CNY"]) // 一个比特币多少钱
		cryptoRate := decimal.NewFromFloat(allRates[currencieItem.Code]) // code(如:code=eth)当前一个比特币能兑换成的多少个eth
		//currencyValuation := fiatRate.DivRound(cryptoRate, 2) // 对应code的每一个数字货币的单价

		var withdraw WithdrawsInfo
		if err := model.DB.Table("withdraws").Select("sum(amount) as amount").Where("currency=? and account_id=?", currencieItem.Code, userID).First(&withdraw).Error; err != nil {
			log.Errorf("Account-获取某个单币种所有的提币-Error: %s", err.Error())
		}

		outAssetsBTC, _ := decimal.NewFromFloat(withdraw.Amount).DivRound(cryptoRate, 8).Float64()
		totalAssetsBTCDecimal := decimal.NewFromFloat(outTotalAssetsBTC).Add(decimal.NewFromFloat(outAssetsBTC))
		outTotalAssetsBTC, _ = totalAssetsBTCDecimal.Round(8).Float64()


		type DepositsInfo struct {
			Amount float64 `json:"amount"`
		}

		var deposit DepositsInfo
		if err := model.DB.Table("deposits").Select("sum(amount) as amount").Where("currency=? and account_id=?", currencieItem.Code, userID).First(&deposit).Error; err != nil {
			log.Errorf("Account-获取某个单币种所有的充值-Error: %s", err.Error())
		}
		inAssetsBTC, _ := decimal.NewFromFloat(deposit.Amount).DivRound(cryptoRate, 8).Float64()
		inTotalAssetsBTCDecimal := decimal.NewFromFloat(inTotalAssetsBTC).Add(decimal.NewFromFloat(inAssetsBTC))
		inTotalAssetsBTC, _ = inTotalAssetsBTCDecimal.Round(8).Float64()

		//fmt.Println(currencieItem.Code)
		//fmt.Println(withdraw.Amount)

	}

	baseSQl := `SELECT %s FROM (SELECT id, created_at, sn, account_id, currency, amount, gas, address, status, done_at, txid, 0 AS type FROM deposits WHERE status != ? UNION ALL SELECT id, created_at, sn, account_id, currency, amount, gas, address, status, done_at, txid, 1 AS type FROM withdraws) AS histories WHERE %v`

	if err := model.DB.Raw(fmt.Sprintf(baseSQl, "histories.*", selectSQL), args...).Order("histories.created_at DESC").Offset(offset).Limit(size).Find(&histories).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Account-Histories-Error: %s", err.Error())
		SendErrJSON("获取明细记录失败", c)
		return
	}

	// 获取记录总数量
	if err := model.DB.Raw(fmt.Sprintf(baseSQl, "COUNT(*)", selectSQL), args...).Count(&totalCount).Error; err != nil {
		log.Errorf("Account-Histories-Error: %s", err.Error())
		SendErrJSON("获取明细记录失败", c)
		return
	}

	// 根据状态ID获取对应的状态说明
	for _, history := range histories {
		if history.Type == depositType {
			history.StatusString = model.DepositStatusDetail(history.Status)
		} else if history.Type == withdrawType {
			history.StatusString = model.WithdrawStatusDetail(history.Status)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"outTotalAssetsBTC": outTotalAssetsBTC,
			"inTotalAssetsBTC":  inTotalAssetsBTC,
			"histories":         histories,
			"pageNo":            page,
			"pageSize":          size,
			"totalPage":         math.Ceil(float64(totalCount) / float64(size)),
			"totalCount":        totalCount,
		},
	})
}

type transferReq struct {
	Receiver string  `json:"receiver" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Note     string  `json:"note"`
}

// 交易转账
func Transfer(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var req transferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("Account-Transfer-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 判断买方和卖方是否是同一用户
	if user.Name == req.Receiver {
		SendErrJSON("接收者不能是发送者", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(req.Currency)
	if err != nil {
		log.Errorf("Account-Transfer-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 获取接收者信息
	var receiver model.User
	if err := model.DB.Model(model.User{}).Where("name = ?", req.Receiver).First(&receiver).Error; err != nil {
		log.Errorf("Account-Transfer-Error: %s", err.Error())
		SendErrJSON("获取接收者信息失败", c)
		return
	}

	// 将小数转换为位
	amountWei := utils.ToWei(req.Amount, int(currency.Precision))

	// 根据当前时间生成流水号
	now := time.Now().Local()
	sn := xid.NewWithTime(now).String()

	transfer := model.Transfer{
		Sn:              sn,
		Currency:        currency.Code,
		Amount:          req.Amount,
		Sender:          user.ID,
		SenderAddress:   user.InternalAddress,
		Receiver:        receiver.ID,
		ReceiverAddress: receiver.InternalAddress,
		Note:            req.Note,
		Status:          model.TransferWaitingInt,
	}

	tx := model.DB.Begin()
	if err := tx.Create(&transfer).Error; err != nil {
		tx.Rollback()
		log.Errorf("Account-Transfer-Error: %s", err.Error())
		return
	}

	// 添加多签名交易，先进行锁币操作
	if _, err := api.PrivateWalletApi.AddTokenTransaction(amountWei.String(), sn, 2, 0, receiver.InternalAddress, user.InternalAddress, currency.PrivateTokenAddress); err != nil {
		tx.Rollback()
		log.Errorf("Account-Transfer-Error: %s", err.Error())
		SendErrJSON("转账失败", c)
		return
	}

	tx.Commit()

	// 将后续执行放币操作放到任务服务中
	go api.TaskApi.AccountTransfer(amountWei.String(), sn, receiver.InternalAddress, user.InternalAddress, currency.PrivateTokenAddress)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
		},
	})
}

var (
	spendType   int32 = 0
	receiveType int32 = 1
)

// 充值/提币记录
type AccountTransfers struct {
	Sn              string     `json:"sn"`               // 交易流水号
	Txid            string     `json:"txid"`             // 交易哈希
	Currency        string     `json:"currency"`         // 代币类型
	Amount          float64    `json:"amount"`           // 转账金额
	Sender          uint64     `json:"sender"`           // 发送者ID
	SenderAddress   string     `json:"receiver_address"` // 接收者地址
	Receiver        uint64     `json:"receiver"`         // 接收者ID
	ReceiverAddress string     `json:"receiver_address"` // 接收者地址
	Status          int32      `json:"status"`           // 状态
	StatusString    string     `json:"status_string"`    // 状态说明
	Note            string     `json:"note"`             // 转账备注
	DoneAt          *time.Time `json:"done_at"`          // 完成时间
	Type            int32      `json:"type"`             // 记录类型
}

// 根据筛选条件获取充值提币记录
func Transfers(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 根据发送还是接收获取对应历史记录，默认为所有
	historyType := c.DefaultQuery("type", "all")

	// 获取代币类型
	code := c.DefaultQuery("code", "all")

	// 获取订单流水号
	Sn := c.Query("sn")

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var (
		selectSQL string
		args      []interface{}
	)

	args = append(args, user.ID, user.ID)
	// 是否进行订单类型筛选
	switch historyType {
	case "all":
		selectSQL = "1 = 1"
	case "spend": // 充值记录
		selectSQL = "transfers.type = ?"
		args = append(args, spendType)
	case "receive": // 提币记录
		selectSQL = "transfers.type = ?"
		args = append(args, receiveType)
	default:
		SendErrJSON("无效的记录类型", c)
		return
	}

	// 是否指定了订单流水号
	if Sn != "" {
		selectSQL += " AND transfers.sn = ?"
		args = append(args, Sn)
	}

	// 是否指定了代币代码
	if code != "all" {
		// 根据代币代码查询代币信息
		currency, err := model.CurrencyFromAndToRedis(strings.ToUpper(code))
		if err != nil {
			SendErrJSON("无效的代币类型", c)
			return
		}

		selectSQL += " AND transfers.currency = ?"
		args = append(args, currency.Code)
	}

	var (
		transfers  []*AccountTransfers
		totalCount int64
	)

	baseSQl := `SELECT %s FROM (SELECT sn, txid, currency, amount, sender, sender_address, receiver, receiver_address, status, note, done_at, created_at, 0 AS type FROM transfers WHERE receiver = ? UNION ALL SELECT sn, txid, currency, amount, sender, sender_address, receiver, receiver_address, status, note, done_at, created_at, 1 AS type FROM transfers WHERE sender = ?) AS transfers WHERE %s`

	if err := model.DB.Raw(fmt.Sprintf(baseSQl, "transfers.*", selectSQL), args...).Order("transfers.done_at DESC").Offset(offset).Limit(size).Find(&transfers).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Account-Transfers-Error: %s", err.Error())
		SendErrJSON("获取转账记录失败", c)
		return
	}

	// 获取记录总数量
	if err := model.DB.Raw(fmt.Sprintf(baseSQl, "COUNT(*)", selectSQL), args...).Count(&totalCount).Error; err != nil {
		log.Errorf("Account-Transfers-Error: %s", err.Error())
		SendErrJSON("获取转账记录失败", c)
		return
	}

	// 根据状态ID获取对应的状态说明
	for _, transfer := range transfers {
		transfer.StatusString = model.TransferStatusDetail(transfer.Status)
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"transfers":  transfers,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}
