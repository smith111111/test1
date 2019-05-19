package model

import (
	"time"
	"errors"

	"github.com/jinzhu/gorm"
)

type Order struct {
	ModelBase                                                                          // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	Sn            string     	`gorm:"index" json:"sn"`                                  // 交易流水号
	Txid          string     	`json:"txid"`                                             // 交易哈希
	Currency      string       	`gorm:"not null; index" json:"currency"`                  // 代币类型
	FiatCurrency  string       	`gorm:"not null; index" json:"fiat_currency"`             // 法币类型
	Price         float64    	`gorm:"not null; type:decimal(32,16)" json:"price"`       // 代币单价
	Amount        float64    	`gorm:"not null; type:decimal(32,16)" json:"amount"`      // 代币数量
	TotalPrice    float64    	`gorm:"not null; type:decimal(32,16)" json:"total_price"` // 代币总价
	Fee           float64    	`gorm:"type:decimal(32,16)" json:"fee"`                   // 手续费
	Buyer         uint64       	`gorm:"not null; index" json:"buyer"`                     // 买方
	BuyerAddress  string     	`json:"buyer_address"`                                    // 买方地址
	Seller        uint64       	`gorm:"not null; index" json:"seller"`                    // 卖方
	SellerAddress string     	`json:"seller_address"`                                   // 卖方地址
	Source        string     	`gorm:"not null; index" json:"source"`                    // 订单来源	广告码
	TradingMethod uint64       	`json:"trading_method"`                                   // 交易方式
	Status        int32        	`gorm:"default: 0" json:"status"`                         // 订单状态	 接单/待付款/待放币/待确认/已完成/待取消/已取消/仲裁中/已仲裁
	CancelBy      uint64       	`json:"cancel_by"`                                        // 订单取消人
	CancelReason  string		`gorm:"size: 150" json:"cancel_reason"`					  // 订单取消原因
	OrderType     int32        	`gorm:"not null; index" json:"order_type"`                // 订单类型 	出售或购买
	DoneAt        *time.Time 	`json:"done_at"`                                          // 完成时间
}

// 根据登录用户的订单角色返回对应的订单类型
func (order *Order) UserOrderType(userID uint64) string {
	switch order.OrderType {
	// 如果是出售订单类型，那么卖方显示是出售，买方显示是购买
	case SellOrderInt:
		if userID == order.Seller {
			return SellOrderString
		} else if userID == order.Buyer {
			return BuyOrderString
		}
		// 如果是购买订单类型，那么买方显示是购买，卖方显示是出售
	case BuyOrderInt:
		if userID == order.Buyer {
			return BuyOrderString
		} else if userID == order.Seller {
			return SellOrderString
		}
	}
	return ""
}

// 订单状态
const (
	OrderWaitingInt            = 0
	OrderWaitingPayInt         = 1
	OrderWaitingReleaseInt     = 2
	OrderWaitingCompleteInt    = 3
	OrderCompletedInt          = 4
	OrderCanceledInt           = 5
	OrderWaitingArbitrationInt = 6
	OrderArbitrationInt        = 7
)

const (
	OrderWaitingString            = "待接单"
	OrderWaitingPayString         = "待付款"
	OrderWaitingReleaseString     = "待放币"
	OrderWaitingCompleteString    = "待确认"
	OrderCompletedString          = "交易完成"
	OrderCanceledString           = "交易取消"
	OrderWaitingArbitrationString = "仲裁中"
	OrderArbitrationString        = "已仲裁"
)

func (order *Order) OrderStatusDetailByAdmin() string {
	switch order.Status {
	case OrderWaitingInt:
		return OrderWaitingString
	case OrderWaitingPayInt:
		return OrderWaitingPayString
	case OrderWaitingReleaseInt:
		return OrderWaitingReleaseString
	case OrderWaitingCompleteInt:
		return OrderWaitingCompleteString
	case OrderCompletedInt:
		return OrderCompletedString
	case OrderCanceledInt:
		return OrderCanceledString
	case OrderWaitingArbitrationInt:
		return OrderWaitingArbitrationString
	case OrderArbitrationInt:
		return OrderArbitrationString
	}
	return ""
}

func (order *Order) OrderStatusDetail(userID uint64) string {
	switch order.Status {
	case OrderWaitingInt:
		if order.OrderType == SellOrderInt { // 如果是出售订单，买家显示待接单，卖家显示已下单
			if userID == order.Seller {
				return "等待买方确认订单"
			} else if userID == order.Buyer {
				return "请及时确认订单"
			}
		} else if order.OrderType == BuyOrderInt { // 如果是购买订单，卖家显示待接单，买家显示已下单
			if userID == order.Seller {
				return "请及时确认订单"
			} else if userID == order.Buyer {
				return "等待卖方确认订单"
			}
		}
	case OrderWaitingPayInt:
		if userID == order.Seller {
			return "等待买方付款"
		} else if userID == order.Buyer {
			return "卖方已确认订单"
		}
	case OrderWaitingReleaseInt:
		if userID == order.Seller {
			return "买方已确认付款"
		} else if userID == order.Buyer {
			return "等待卖方确认放币"
		}
	case OrderWaitingCompleteInt:
		if userID == order.Seller {
			return "等待买方确认收币"
		} else if userID == order.Buyer {
			return "卖方已确认放币"
		}
	case OrderCompletedInt:
		return "交易完成"
	case OrderCanceledInt:
		return "交易取消"
	case OrderWaitingArbitrationInt:
		return "交易仲裁中"
	case OrderArbitrationInt:
		return "已仲裁"
	}
	return ""
}

// 订单类型
const (
	SellOrderInt = 0
	BuyOrderInt  = 1
)

const (
	SellOrderString = "出售"
	BuyOrderString  = "购买"
)

func (order *Order) OrderTypeDetailByAdmin() string {
	switch order.OrderType {
	case SellOrderInt:
		return SellOrderString
	case BuyOrderInt:
		return BuyOrderString
	}
	return ""
}

// 订单信息
type OrderInfo struct {
	Sn               string         `json:"sn"`                 // 交易流水号
	Txid             string         `json:"txid"`               // 交易哈希
	TraderID       	 uint64         `json:"trader_id"`        	// 交易人ID
	Trader           string         `json:"trader"`             // 交易人名称
	Price            float64        `json:"price"`              // 代币单价
	Amount           float64        `json:"amount"`             // 代币数量
	TotalPrice       float64        `json:"total_price"`        // 代币总价
	Fee              float64        `json:"fee"`                // 手续费
	Status           string         `json:"status"`             // 订单状态
	OrderType        string         `json:"order_type"`         // 订单类型
	CountDownSeconds int32          `json:"count_down_seconds"` // 倒计时
	Offer            OrderOfferInfo `json:"offer"`              // 广告信息
}

// 初始化并赋值
func (orderInfo *OrderInfo) Init(currentPrice float64, user *User, offer *Offer, order *Order, orderEvent *OrderEvent) error {
	orderInfo.Sn = order.Sn
	orderInfo.Txid = order.Txid
	orderInfo.Price = order.Price
	orderInfo.Amount = order.Amount
	orderInfo.TotalPrice = order.TotalPrice
	orderInfo.Fee = order.Fee
	orderInfo.OrderType = order.UserOrderType(user.ID)

	var traderID uint64
	// 如果用户是卖方，那么交易人就是买方，如果用户是买方，交易人就是卖方
	if order.Seller == user.ID {
		traderID = order.Buyer
	} else if order.Buyer == user.ID {
		traderID = order.Seller
	} else {
		return errors.New("您没有权限执行此操作")
	}

	// 获取交易人的信息
	var trader User
	if err := DB.First(&trader, traderID).Error; err != nil {
		return errors.New("获取交易人信息失败")
	}
	orderInfo.Trader = trader.Name
	orderInfo.TraderID = trader.ID

	if orderEvent == nil {
		var newOrderEvent OrderEvent
		if err := DB.Where("order_id = ? AND status = ?", order.ID, order.Status).First(&newOrderEvent).Order("created_at DESC").Error; err == nil {
			t1 := time.Now().Local()
			t2 := newOrderEvent.InvalidTime

			if t1.Before(t2) {
				subM := t2.Sub(t1)
				orderInfo.CountDownSeconds = int32(subM.Seconds())
			}
		}
		orderEvent = &newOrderEvent
	} else {
		t1 := time.Now().Local()
		t2 := orderEvent.InvalidTime

		if t1.Before(t2) {
			subM := t2.Sub(t1)
			orderInfo.CountDownSeconds = int32(subM.Seconds())
		}
	}

	if orderInfo.CountDownSeconds == 0 {
		// 倒计时为０时交易中的订单都改为已取消,待确认的则标记为已确认
		switch order.Status {
		case OrderWaitingInt, OrderWaitingPayInt, OrderWaitingReleaseInt:
			doneAt := time.Now().Local()
			if err := DB.Model(&order).Updates(Order{Status: OrderCanceledInt, DoneAt: &doneAt, CancelReason: "订单超时取消"}).Error; err != nil {
				return errors.New("服务器错误")
			}
		case OrderWaitingCompleteInt:
			doneAt := time.Now().Local()
			if err := DB.Model(&order).Updates(Order{Status: OrderCompletedInt, DoneAt: &doneAt}).Error; err != nil {
				return errors.New("服务器错误")
			}
		}
	}

	orderInfo.Status = order.OrderStatusDetail(user.ID)

	offerInfo := OrderOfferInfo{
		Code:             order.Source,
		Currency:     offer.Currency,
		FiatCurrency: offer.FiatCurrency,
		Price:            currentPrice,
		MinLimit:         offer.MinLimit,
		MaxLimit:         offer.MaxLimit,
	}

	// 获取广告主名称
	if trader.ID == offer.AccountID {
		offerInfo.Publisher = trader.Name
		offerInfo.PublisherID = trader.ID
	} else if user.ID == offer.AccountID {
		offerInfo.Publisher = user.Name
		offerInfo.PublisherID = user.ID
	}

	var (
		selectSQL string
		args      []interface{}
	)

	switch order.OrderType {
	// 如果是出售订单, 广告者是买方
	case SellOrderInt:
		selectSQL = "buyer = ?"
		args = append(args, offer.AccountID)
		// 如果是购买订单, 广告者是卖方
	case BuyOrderInt: // 购买订单
		selectSQL = "seller = ?"
		args = append(args, offer.AccountID)
	}

	selectSQL += " AND currency = ? AND fiat_currency = ? AND status = ?"
	args = append(args, offer.Currency, offer.FiatCurrency, OrderCompletedInt)

	// 获取广告主的订单成交次数
	if err := DB.Model(Order{}).Where(selectSQL, args...).Count(&offerInfo.OrderCount).Error; err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("获取用户成交次数失败")
	}

	// TODO 获取广告主的被仲裁成功订单数
	var arbitrationCount uint
	if err := DB.Model(Arbitration{},
	).Joins("JOIN orders ON arbitrations.order_id = orders.id",
	).Where("arbitrations.arbitration_result = ? AND (orders.seller = ? OR orders.buyer = ?)", ArbitrationSuccessful, offer.AccountID, offer.AccountID).Count(&arbitrationCount).Error; err != nil {
		return errors.New("计算用户好评度失败")
	}

	offerInfo.Popularity = 100
	// 好评度等于 100 - (仲裁订单数 / 总订单数 * 100)
	if arbitrationCount != 0 && offerInfo.OrderCount != 0 {
		offerInfo.Popularity = 100 - (float64(arbitrationCount) / float64(offerInfo.OrderCount) * 100)
	}

	orderInfo.Offer = offerInfo

	return nil
}

// 后台订单基本信息
type OrderAdminBaseInfo struct {
	CreatedAt        	time.Time 	`json:"created_at"`         	// 创建时间
	Sn               	string    	`json:"sn"`                 	// 交易流水号
	Price            	float64   	`json:"price"`              	// 代币单价
	Amount           	float64   	`json:"amount"`             	// 代币数量
	TotalPrice       	float64   	`json:"total_price"`        	// 代币总价
	Fee              	float64   	`json:"fee"`                	// 手续费
	StatusInt        	int32       `json:"status_int"`         	// 订单状态
	Status           	string    	`json:"status"`             	// 订单状态详情
	OrderTypeInt        int32    	`json:"order_type_int"`         // 订单类型
	OrderType        	string    	`json:"order_type"`         	// 订单类型详情
	Currency     		string    	`json:"currency"`      			// 代币类型
	FiatCurrency 		string    	`json:"fiat_currency"` 			// 法币类型
	Buyer        		User       	`json:"buyer"`         			// 买家
	Seller       		User       	`json:"seller"`        			// 卖家
}

func (baseInfo *OrderAdminBaseInfo) Init(order *Order) error {
	baseInfo.CreatedAt = order.CreatedAt
	baseInfo.Sn = order.Sn
	baseInfo.Price = order.Price
	baseInfo.Amount = order.Amount
	baseInfo.TotalPrice = order.TotalPrice
	baseInfo.Fee = order.Fee
	baseInfo.StatusInt = order.Status
	baseInfo.Status = order.OrderStatusDetailByAdmin()
	baseInfo.OrderTypeInt = order.OrderType
	baseInfo.OrderType = order.OrderTypeDetailByAdmin()

	//根据广告码获取广告信息
	var offer Offer

	if err := DB.Table("offers").Where("code = ?", order.Source).First(&offer).Error; err != nil {
		return errors.New("获取广告详情失败")
	}

	// 获取买方的信息
	if err := DB.First(&baseInfo.Buyer, order.Buyer).Error; err != nil {
		return errors.New("获取买方信息失败")
	}

	// 获取卖方的信息
	if err := DB.First(&baseInfo.Seller, order.Seller).Error; err != nil {
		return errors.New("获取卖方信息失败")
	}

	baseInfo.Currency = offer.Currency
	baseInfo.FiatCurrency = offer.FiatCurrency
	return nil
}

// 后台订单详细信息
type OrderAdminDetailInfo struct {
	CreatedAt    time.Time    `json:"created_at"`    // 创建时间
	Sn           string       `json:"sn"`            // 交易流水号
	Txid         string       `json:"txid"`          // 交易哈希
	Price        float64      `json:"price"`         // 代币单价
	Amount       float64      `json:"amount"`        // 代币数量
	TotalPrice   float64      `json:"total_price"`   // 代币总价
	Fee          float64      `json:"fee"`           // 手续费
	StatusInt    int32         `json:"status_int"`    // 订单状态
	Status       string       `json:"status"`        // 订单状态详情
	OrderType    string       `json:"order_type"`    // 订单类型
	Buyer        User         `json:"buyer"`         // 买家
	Seller       User         `json:"seller"`        // 卖家
	Currency     Currency     `json:"currency"`      // 代币信息
	FiatCurrency FiatCurrency `json:"fiat_currency"` // 法币信息
	Source       string       `json:"source"`        // 订单来源
}

func (detailInfo *OrderAdminDetailInfo) Init(currentPrice float64, order *Order, offer *Offer, tradingMethodMap map[uint64]*TradingMethod) error {
	detailInfo.CreatedAt = order.CreatedAt
	detailInfo.Sn = order.Sn
	detailInfo.Txid = order.Txid
	detailInfo.Price = order.Price
	detailInfo.Amount = order.Amount
	detailInfo.TotalPrice = order.TotalPrice
	detailInfo.Fee = order.Fee
	detailInfo.StatusInt = order.Status
	detailInfo.Status = order.OrderStatusDetailByAdmin()
	detailInfo.OrderType = order.OrderTypeDetailByAdmin()
	detailInfo.Currency, _ = CurrencyFromAndToRedis(offer.Currency)
	detailInfo.FiatCurrency, _ = FiatCurrencyFromAndToRedis(offer.FiatCurrency)
	detailInfo.Source = order.Source

	// 获取买方的信息
	if err := DB.First(&detailInfo.Buyer, order.Buyer).Error; err != nil {
		return errors.New("获取买方信息失败")
	}

	// 获取卖方的信息
	if err := DB.First(&detailInfo.Seller, order.Seller).Error; err != nil {
		return errors.New("获取卖方信息失败")
	}
	return nil
}

// 获取最近交易记录
type RecentlyOrder struct {
	Code             string    `json:"code"`
	OrderType        int32       `json:"order_type"`
	Currency     	 string    `json:"currency_code"`
	FiatCurrency 	 string    `json:"fiat_currency_code"`
	Amount           float64   `json:"amount"`
	TotalPrice       float64   `json:"total_price"`
	DoneAt           time.Time `json:"done_at"`
}