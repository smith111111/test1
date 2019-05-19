package model

import "time"

type Transfer struct {
	Sn        			string     		`gorm:"index" json:"sn"`                    	// 交易流水号
	Txid      			string     		`json:"txid"`                              		// 交易哈希
	Currency  			string      	`json:"currency"`             					// 代币类型
	Amount    			float64    		`gorm:"type:decimal(32,16)" json:"amount"` 		// 转账金额
	Sender	 			uint64      	`gorm:"index" json:"sender"`           			// 发送者ID
	SenderAddress   	string     		`json:"receiver_address"`              			// 接收者地址
	Receiver			uint64			`gorm:"index" json:"receiver"`					// 接收者ID
	ReceiverAddress   	string     		`json:"receiver_address"`              			// 接收者地址
	Status    			int32        	`gorm:"default:0" json:"status"`           		// 状态
	Note      			string     		`json:"note"`                              		// 转账备注
	CreatedAt 			time.Time 		`json:"created_at"`								// 创建时间
	DoneAt    			*time.Time 		`json:"done_at"`                           		// 完成时间
}

// 订单状态
const (
	TransferWaitingInt            	= 0
	TransferSuccessInt            	= 1
	TransferFailedInt         		= 2
)

const (
	TransferWaitingString        = "转账中"
	TransferSuccessString        = "转账成功"
	TransferFailedString         = "转账失败"
)

func TransferStatusDetail(statusInt int32) (statusString string) {
	switch statusInt {
	case TransferWaitingInt:
		statusString = TransferWaitingString
	case TransferSuccessInt:
		statusString = TransferSuccessString
	case TransferFailedInt:
		statusString = TransferFailedString
	}
	return
}