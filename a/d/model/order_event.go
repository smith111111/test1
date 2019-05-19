package model

import "time"

type OrderEvent struct {
	ModelBase                                                        	// fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	OrderID     uint64      `gorm:"not null" json:"order_id"`          	// 订单编号
	Status      int32       `gorm:"default: 0" json:"status"`          	// 订单状态	待接单/交易中/交易完成
	OrderType   int32       `gorm:"not null; index" json:"order_type"` 	// 订单类型 	出售或购买
	Txid      	string    	`json:"txid"`                            	// 交易哈希
	InvalidTime time.Time 	`json:"invalid_time"`                      	// 失效时间
	TurnTime    time.Time 	`json:"turn_time"`                         	// 状态反转时间
}