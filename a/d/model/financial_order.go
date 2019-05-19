package model

import "time"

type FinancialOrder struct {
	ID                  uint64    `gorm:"primary_key" json:"id"`
	Name                string    `json:"name"`                         // 产品名称
	Buyer               uint64    `gorm:"not null; index" json:"buyer"` // 买方
	ProductID           uint64    `json:"product_id"`                   // 产品的id
	ShowEarningsPercent string    `json:"show_earnings_percent"`        // 前端显示收益百分比
	ProductCycle        int32     `json:"product_cycle"`                // 产品周期
	BuyAmount           float64   `gorm:"type:decimal(32,8)" json:"buy_amount"`                   // 认购的数量
	PredictEarnings     float64   `gorm:"type:decimal(32,8)" json:"predict_earnings"`             // 预计收益
	RealTotalEarnings   float64   `gorm:"type:decimal(32,8)" json:"real_total_earnings"`          // 实际总收益
	Status              int32     `json:"status"`                       // 状态  0=锁仓中,1=锁仓计息中,2=已派息
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Remark              string    `json:"remark"`
}

type FinancialOrderProduct struct {
	FinancialOrder   FinancialOrder   `json:"financial_order"`
	FinancialProduct FinancialProduct `json:"financial_product"`
	EndBuy           string           `json:"endBuy"`
}

// 订单状态
const (
	FinancialOrderLockInt           = 0
	FinancialOrderInterestIngInt    = 1
	FinancialOrderSendedInterestInt = 2
)

const (
	FinancialOrderLockString           = "锁仓中"
	FinancialOrderInterestIngString    = "锁仓计息中"
	FinancialOrderSendedInterestString = "已派息"
)
