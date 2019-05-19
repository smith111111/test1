package model

import "time"

type FinancialProduct struct {
	ID                  uint64    `gorm:"primary_key" json:"id"`
	Name                string    `json:"name"`                                      // 产品名称
	BitCoin             string    `json:"bit_coin"`                                  // 数字币种
	ShowEarningsPercent string    `json:"show_earnings_percent"`                     // 前端显示收益百分比 用String类型是 有一种情况有个区间值
	RealEarningsPercent float64   `json:"real_earnings_percent"`                     // 后台显示收益百分比 用于计算当天的收益
	ProductCycle        int32     `json:"product_cycle"`                             // 产品周期
	MinCastAmount       float64   `gorm:"type:decimal(32,8)" json:"min_cast_amount"` // 最小起投量
	TotalRaise          float64   `gorm:"type:decimal(32,8)" json:"total_raise"`     // 当前标的总量
	Label               string    `json:"label"`                                     // 标签 多个以','分割
	Status              int32     `json:"status"`                                    // 状态 0=募集中,1=标满,2=计息中,3=已结束
	IsDelete            int32     `json:"is_delete"`                                 // 是否下架
	StartRaise          time.Time `json:"start_raise"`                               // 开始募集时间
	EndBuy              time.Time `json:"end_buy"`                                   // 认购截止时间
	StartInterest       time.Time `json:"start_interest"`                            // 起息时间
	ExpireTime          time.Time `json:"expire_time"`                               // 到期时间
	SendInterest        time.Time `json:"send_interest"`                             // 返息时间
	EarningsMethod      string    `json:"earnings_method"`                           // 收益方式
	RuleRemark          string    `json:"rule_remark"`                               // 规则
	BuyAmount           float64   `gorm:"type:decimal(32,8)" json:"buy_amount"`      // 认购的数量mingn 冗余字段 方便计算查询
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	Remark              string    `gorm:"size:4000" json:"remark"`
}

// 产品状态
const (
	FinancialProductRaiseInt     = 0
	FinancialProductFullScaleInt = 1
	FinancialProductIssueIngInt  = 2
	FinancialProductIssueEndInt  = 3
)

const (
	FinancialProductRaiseString     = "募集中"
	FinancialProductFullScaleString = "标满"
	FinancialProductIssueIngString  = "计息中"
	FinancialProductIssueEndString  = "已结束"
)
