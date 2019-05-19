package model

import "time"

type FinancialEarningsHis struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	BitCoin   string    `json:"bit_coin"`                     // 数字币种
	Name      string    `json:"name"`                         // 产品名称
	Buyer     uint64    `gorm:"not null; index" json:"buyer"` // 买方
	ProductID uint64    `json:"product_id"`                   // 产品的id
	Earnings  float64   `gorm:"type:decimal(32,8)" json:"earnings"`                     // 今天收益
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
