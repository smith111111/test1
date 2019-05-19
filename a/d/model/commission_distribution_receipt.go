package model

import "time"

type CommissionDistributionReceipt struct {
	ModelBase                                                                // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	StartAt  time.Time 		`json:"start_at"`                                    // 开始时间
	EndAt    time.Time 		`json:"end_at"`                                      // 结束时间
	UserID   uint64       	`gorm:"not null; index" json:"user_id"`              // 用户编号
	Currency string       	`gorm:"not null; index" json:"currency"`             // 代币类型
	Amount   float64    	`gorm:"not null; type:decimal(32,16)" json:"amount"` // 代币数量
	Status   int32        	`gorm:"default: 1" json:"status"`                    // 状态		成功/失败
	Remark   string     	`json:"remark"`                                      // 备注
}


const (
	// 成功
	CommissionDistributionReceipt_Success = 1
	// 失败
	CommissionDistributionReceipt_Fail = 2
)