package model

import (
	"time"
)

type Financial struct {
	BitCoin           string    `json:"bit_coin"` // 数字币种
	Buyer             uint64    `json:"buyer"`                                 // 购买人
	LockAmount        float64   `gorm:"type:decimal(32,8)" json:"lock_amount"`                           // 锁仓数量
	YesterdayEarnings float64   `gorm:"type:decimal(32,8)" json:"yesterday_earnings"`                    // 昨日收益
	TotalEarnings     float64   `gorm:"type:decimal(32,8)" json:"total_earnings"`                        // 累计收益
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Remark            string    `json:"remark"`
}

//func (financial *Financial) BeforeUpdate(scope *gorm.Scope) (err error) {
//	endTime := utils.GetTodayTime().Format("2006-01-02")
//	fmt.Println(endTime)
//	fmt.Println(financial.UpdatedAt)
//	//if financial.UpdatedAt.After(endTime) {
//		fmt.Println("QUANXIAO")
//		scope.SetColumn("yesterday_earnings", 0.0)
//	//}
//	return nil
//}
