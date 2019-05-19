package model

import "time"

type CommissionDistribution struct {
	ModelBase                                                                          		// fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	Sn            string     	`gorm:"index" json:"sn"`                                  	// 流水号
	Txid          string     	`json:"txid"`                                             	// 交易哈希
	Currency      string       	`gorm:"not null; index" json:"currency"`                  	// 代币类型
	Rate          float64    	`gorm:"not null; type:decimal(32,16)" json:"rate"`        	// 分佣比例
	Amount        float64    	`gorm:"not null; type:decimal(32,16)" json:"amount"`      	// 代币数量
	UserID        uint64       	`gorm:"not null; index" json:"user_id"`                   	// 用户编号
	UserType      int32       	`gorm:"default: 0" json:"user_type"`                      	// 用户类型	普通会员/天使合伙人/创世合伙人
	OrderID       uint64       	`gorm:"not null; index" json:"order_id"`                  	// 订单编号
	BusinessType  int32  	 	`gorm:"default: 0" json:"business_type"`				   	// 类型       OTC/Others
	Status        int32        	`gorm:"default: 1" json:"status"`                         	// 状态		草稿/发放成功/发放失败
	DoneAt        *time.Time 	`json:"done_at"`                                          	// 完成时间
}

type CommissionRate struct {
	UserID        uint64
	UserType      int32
	Rate          float64
}

type CommissionUser struct {
	UserID        uint64
	UserType      int32
	Status		  int32
	Level   	  int32
}

type CommissionInfo struct {
	BusinessType  	int32     	`json:"business_type"`
	Amount 			float64 	`json:"amount"`
	CreatedAt		time.Time	`json:"created_at"`
}

type DistributionCommissionInfo struct {
	UserID        			uint64		`json:"user_id"`
	Currency      			string		`json:"currency"`
	Amount        			float64		`json:"amount"`
	Precision				uint		`json:"precision"`
	InternalAddress			string		`json:"internal_address"`
	PrivateTokenAddress		string		`json:"private_token_address"`
}

const (
	// 草稿
	CommissionDistributionStatus_Draft int32 = 0
	// 发放成功
	CommissionDistributionStatus_Success int32 = 1
	// 发放失败
	CommissionDistributionStatus_Fail int32 = 2
)

const (
	// 直接上级佣金比例
	CommissionRate_Parent = 0.3
	// 合伙人提成佣金比例
	CommissionRate_Partnership = 0.05
)