package model

import (
	"fmt"
	"time"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type Deposit struct {
	ModelBase
	Sn            string     	`gorm:"index" json:"sn"`                                    // 交易流水号
	AccountID     uint64     	`json:"account_id"`                                         // 账户ID
	Currency      string     	`json:"currency"`                                           // 代币类型
	Amount        float64    	`gorm:"type:decimal(32,16)" json:"amount"`                  // 充值金额
	Gas           float64    	`gorm:"type:decimal(32,16)" json:"gas"`                     // 矿工费
	Address       string     	`gorm:"index" json:"address"`                               // 充值地址
	Status        int32        	`json:"status"`                                             // 状态
	DoneAt        *time.Time 	`json:"done_at"`                                            // 完成时间
	Txid          string     	`json:"txid"` 												// 交易哈希
	Memo          string     	`json:"memo"`                                               // 助记词
	Confirmations string     	`json:"confirmations"`                                      // 确认数
}

// 充值信息
type DepositInfo struct {
	ID            uint64    `json:"id"`            //ID编号
	Sn            string    `json:"sn"`            // 交易流水号
	Txid          string    `json:"txid"`          // 交易哈希
	UserName      string    `json:"user_name"`     // 用户名
	Amount        float64   `json:"amount"`        // 提币金额
	Gas           float64   `json:"gas"`           // 矿工费
	Address       string    `json:"address"`       // 目标地址
	Status        int32     `json:"status"`        // 状态码
	StatusString  string    `json:"status_string"` // 状态详情
	DoneAt        time.Time `json:"done_at"`       // 完成时间
	CreatedAt     time.Time `json:"created_at"`    //创建时间
	Currency      string    `json:"currency"`      // 代币类型
	Memo          string    `json:"memo"`          // 提币备注
	Confirmations string    `json:"confirmations"` // 确认数
}

const (
	DepositNotInt       = 0
	DepositPendingInt   = 1
	DepositCompletedInt = 2
	DepositAbnormalInt  = 3
)

const (
	DepositNotString       = "未充值"
	DepositPendingString   = "充值中"
	DepositCompletedString = "已充值"
	DepositAbnormalString  = "充值异常"
)

func DepositStatusDetail(statusInt int32) (statusString string) {
	switch statusInt {
	case DepositNotInt:
		statusString = DepositNotString
	case DepositPendingInt:
		statusString = DepositPendingString
	case DepositCompletedInt:
		statusString = DepositCompletedString
	case DepositAbnormalInt:
		statusString = DepositAbnormalString
	}
	return
}

// DepositFromRedis 从redis中取出用户未使用的充值地址
func DepositFromRedis(address string) (Deposit, error) {
	depositKey := fmt.Sprintf("%s%s", DepositAddress, address)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	var deposit Deposit

	depositBytes, err := redis.Bytes(RedisConn.Do("GET", depositKey))
	if err != nil {
		// 获取失败，再尝试从数据库中查询
		if err := DB.Where("address = ? AND (status = ? OR status = ?)", address, DepositNotInt, DepositPendingInt).First(&deposit).Error; err != nil {
			return Deposit{}, err
		}
		return deposit, nil
	}

	bytesErr := json.Unmarshal(depositBytes, &deposit)
	if bytesErr != nil {
		return Deposit{}, err
	}
	return deposit, nil
}

// DepositToRedis 将用户未使用的充值地址存到redis
func DepositToRedis(deposit Deposit) error {
	depositBytes, err := json.Marshal(deposit)
	if err != nil {
		return err
	}
	depositKey := fmt.Sprintf("%s%s", DepositAddress, deposit.Address)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	if _, redisErr := RedisConn.Do("SET", depositKey, depositBytes); redisErr != nil {
		return err
	}
	return nil
}

// DepositDeleteRedis 将用户已使用的充值地址从redis中删除
func DepositDeleteRedis(address string) error {
	depositKey := fmt.Sprintf("%s%s", DepositAddress, address)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	_, err := RedisConn.Do("DEL", depositKey)
	if err != nil {
		return err
	}

	return nil
}

// Excel提币信息
type DepositExportInfo struct {
	ID           uint64  `json:"id"`            // 交易ID
	Sn           string  `json:"sn"`            // 交易流水号
	Txid         string  `json:"txid"`          // 交易哈希
	Amount       float64 `json:"amount"`        // 提币金额
	Gas          float64 `json:"gas"`           // 矿工费
	Address      string  `json:"address"`       // 目标地址
	Status       int32   `json:"status"`        // 状态码
	StatusString string  `json:"status_string"` // 状态详情
	DoneAt       string  `json:"done_at"`       // 完成时间
	CurrencyCode string  `json:"currency_code"` // 代币代码
	Memo         string  `json:"memo"`          // 提币备注
	UserName     string  `json:"user_name"`     // 充值用户
	CreatedAt    string  `json:"created_at"`    // 创建时间
}
