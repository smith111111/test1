package model

import (
	"errors"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// 出售/购买广告
type Offer struct {
	ModelBase                                                                    // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	OfferType        	int32     	`gorm:"not null; index" json:"offer_type"`          // 广告类型 出售或购买
	TradingType      	int32     	`gorm:"not null" json:"trading_type"`               // 交易类型 钱包或账户
	Currency     		string  	`gorm:"size:10" json:"currency_code"`               // 代币代码
	FiatCurrency 		string  	`gorm:"size:10" json:"fiat_currency_code"`          // 法币代码
	Price            	float64 	`gorm:"not null; type:decimal(32,16)" json:"price"` // 价格
	Float            	float64 	`gorm:"type:decimal(32,16)" json:"float"`           // 溢价浮动
	AcceptPrice      	float64 	`gorm:"type:decimal(32,16)" json:"accept_price"`    // 可接受价格 最低出售价格或最高购买价格
	TradingMethods   	string  	`gorm:"not null" json:"trading_methods"`            // 交易方式
	MinLimit         	float64 	`gorm:"not null" json:"min_limit"`                  // 最小限额
	MaxLimit         	float64 	`gorm:"not null" json:"max_limit"`                  // 最大限额
	Note             	string  	`gorm:"size:200" json:"note"`                       // 交易备注
	Status           	int32     	`gorm:"default:0" json:"status"`                    // 广告状态
	AccountID        	uint64  	`gorm:"index" json:"account_id"`                    // 发布者
	Code             	string  	`gorm:"size:50" json:"code"`                        // 广告码
}

// 广告类型
const (
	BuyOfferInt  = 0
	SellOfferInt = 1
)

const (
	BuyOfferString  = "购买广告"
	SellOfferString = "出售广告"
)

func (offer *Offer) OfferTypeString() (typeString string) {
	switch offer.OfferType {
	case BuyOfferInt:
		typeString = BuyOfferString
	case SellOfferInt:
		typeString = SellOfferString
	}
	return
}

// 交易类型
const (
	AccountTradingInt = 0
	WalletTradingInt  = 1
)

const (
	AccountTradingString = "账户交易"
	WalletTradingString  = "钱包交易"
)

func (offer *Offer) OfferTradingTypeString() (tradingTypeString string) {
	switch offer.TradingType {
	case AccountTradingInt:
		tradingTypeString = AccountTradingString
	case WalletTradingInt:
		tradingTypeString = WalletTradingString
	}
	return
}

const (
	OfferOnInt  = 0
	OfferOffInt = 1
)

const (
	OfferOnString  = "已上架"
	OfferOffString = "已下架"
)

func (offer *Offer) OfferStatusString() (statusString string) {
	switch offer.Status {
	case OfferOnInt:
		statusString = OfferOnString
	case OfferOffInt:
		statusString = OfferOffString
	}
	return
}

// 广告信息
type OfferInfo struct {
	Code             string           	`json:"code"`               // 广告码
	OfferTypeInt     int32              `json:"offer_type_int"`     // 广告类型
	OfferType        string           	`json:"offer_type"`         // 广告类型详情
	TradingTypeInt   int32              `json:"trading_type_int"`   // 交易类型
	TradingType      string           	`json:"trading_type"`       // 交易类型详情
	CurrencyCode     string           	`json:"currency_code"`      // 代币代码
	FiatCurrencyCode string           	`json:"fiat_currency_code"` // 法币代码
	Price            float64          	`json:"price"`              // 价格
	MinLimit         float64          	`json:"min_limit"`          // 最小限额
	MaxLimit         float64          	`json:"max_limit"`          // 最大限额
	TradingMethods   []*TradingMethod 	`json:"trading_methods"`    // 交易方式
	Note             string           	`json:"note"`               // 交易备注
	StatusInt        int32              `json:"status_int"`         // 广告状态
	Status           string           	`json:"status"`             // 广告状态详情
	Publisher        string           	`json:"publisher"`          // 发布者
	PublisherID    	 uint64           	`json:"publisher_id"`     	// 发布者ID
	OrderCount       uint32             `json:"order_count"`        // 订单成交量
	Popularity       float64          	`json:"popularity"`         // 好评度
	AcceptPrice      float64          	`json:"accept_price"`       // 可接受价格
	Float            float64          	`json:"float"`              // 溢价浮动
}

// 初始化并赋值
func (offerInfo *OfferInfo) Init(currentPrice float64, offer *Offer, tradingMethodMap map[uint64]*TradingMethod) error {
	offerInfo.Code = offer.Code
	offerInfo.CurrencyCode = offer.Currency
	offerInfo.FiatCurrencyCode = offer.FiatCurrency
	offerInfo.MinLimit = offer.MinLimit
	offerInfo.MaxLimit = offer.MaxLimit
	offerInfo.AcceptPrice = offer.AcceptPrice
	offerInfo.Note = offer.Note
	offerInfo.Float = offer.Float
	offerInfo.OfferTypeInt = offer.OfferType
	offerInfo.OfferType = offer.OfferTypeString()
	offerInfo.TradingTypeInt = offer.TradingType
	offerInfo.TradingType = offer.OfferTradingTypeString()
	offerInfo.StatusInt = offer.Status
	offerInfo.Status = offer.OfferStatusString()

	// 根据当前汇率计算价格
	offerInfo.Price, _ = decimal.NewFromFloat(currentPrice).Mul(decimal.NewFromFloat(offer.Float)).Round(2).Float64()

	// 如果用户设置了可接受价格，则判断当前价格是否可接受
	if offer.OfferType == BuyOfferInt {
		// 如果是购买广告则判断当前价格是否高于可接受价格
		if offer.AcceptPrice != 0 && offerInfo.Price > offer.AcceptPrice {
			offerInfo.Price = offer.AcceptPrice
		}
	} else if offer.OfferType == SellOfferInt {
		// 如果是出售广告则判断当前价格是否低于可接受价格
		if offer.AcceptPrice != 0 && offerInfo.Price < offer.AcceptPrice {
			offerInfo.Price = offer.AcceptPrice
		}
	}

	// 获取广告主的个人信息
	var publisher User
	if err := DB.Table("users").Where("id = ?", offer.AccountID).First(&publisher).Error; err != nil {
		return errors.New("获取广告主的个人信息失败")
	}
	offerInfo.Publisher = publisher.Name
	offerInfo.PublisherID = publisher.ID

	var (
		selectSQL string
		args      []interface{}
	)

	switch offer.OfferType {
	// 如果是购买广告, 广告者是买方
	case BuyOfferInt:
		selectSQL = "buyer = ?"
		args = append(args, offer.AccountID)
		// 如果是出售广告, 广告者是卖方
	case SellOfferInt:
		selectSQL = "seller = ?"
		args = append(args, offer.AccountID)
	}

	selectSQL += " AND currency = ? AND fiat_currency = ? AND status = ?"
	args = append(args, offer.Currency, offer.FiatCurrency, OrderCompletedInt)

	// 获取广告主的订单成交次数
	if err := DB.Model(Order{}).Where(selectSQL, args...).Count(&offerInfo.OrderCount).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("获取广告主的订单成交次数失败")
	}

	// TODO 获取广告主的被仲裁成功订单数
	var arbitrationCount uint
	if err := DB.Model(Arbitration{},
	).Joins("JOIN orders ON arbitrations.order_id = orders.id",
	).Where("arbitrations.arbitration_result = ? AND (orders.seller = ? OR orders.buyer = ?)", ArbitrationSuccessful, offer.AccountID, offer.AccountID).Count(&arbitrationCount).Error; err != nil {
		return errors.New("获取广告主的仲裁订单数失败")
	}

	offerInfo.Popularity = 100
	// 好评度等于 100 - (仲裁订单数 / 总订单数 * 100)
	if arbitrationCount != 0 && offerInfo.OrderCount != 0 {
		offerInfo.Popularity = 100 - (float64(arbitrationCount) / float64(offerInfo.OrderCount) * 100)
	}

	for _, methodId := range strings.Split(offer.TradingMethods, ",") {
		id, _ := strconv.ParseUint(methodId, 10, 64)
		offerInfo.TradingMethods = append(offerInfo.TradingMethods, tradingMethodMap[uint64(id)])
	}
	return nil
}

// 订单明细广告信息
type OrderOfferInfo struct {
	Code             string  	`json:"code"`               // 广告码
	Currency     	 string  	`json:"currency"`      		// 代币代码
	FiatCurrency 	 string  	`json:"fiat_currency"` 		// 法币代码
	Price            float64 	`json:"price"`              // 价格
	MinLimit         float64 	`json:"min_limit"`          // 最小限额
	MaxLimit         float64 	`json:"max_limit"`          // 最大限额
	Publisher        string  	`json:"publisher"`          // 发布者
	OrderCount       uint32   	`json:"order_count"`        // 订单成交量
	Popularity       float64 	`json:"popularity"`         // 好评度
	PublisherID    	 uint64  	`json:"publisher_id"`     	 // 发布者ID
}

// 广告列表, 实现sort包排序方法
type OfferList []*OfferInfo

// 获取长度
func (o OfferList) Len() int {
	return len(o)
}

// 根据广告金额升序
func (o OfferList) Less(i, j int) bool {
	return o[i].Price > o[j].Price
}

// 交换数据
func (o OfferList) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}