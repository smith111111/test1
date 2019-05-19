package model

import (
	"fmt"
	"time"
	"encoding/json"

	"github.com/garyburd/redigo/redis"
)

type Withdraw struct {
	// 提币
	ModelBase
	Sn        string     	`gorm:"index" json:"sn"`                    // 交易流水号
	AccountID uint64      	`gorm:"index" json:"account_id"`           	// 账户ID
	Currency  string       	`json:"currency"`             				// 代币类型
	Amount    float64    	`gorm:"type:decimal(32,16)" json:"amount"` 	// 提现金额
	Fee       float64    	`gorm:"type:decimal(32,16)" json:"fee"`   	// 手续费
	Gas       float64    	`gorm:"type:decimal(32,16)" json:"gas"`    	// 矿工费
	Address   string     	`json:"address"`              				// 目标地址
	Status    int32        	`gorm:"default:0" json:"status"`           	// 状态
	DoneAt    *time.Time 	`json:"done_at"`                           	// 完成时间
	Txid      string     	`json:"txid"`                              	// 交易哈希
	Memo      string     	`json:"memo"`                              	// 助记词
	Note      string     	`json:"note"`                              	// 提币备注
}

const (
	WithdrawPendingInt   = 1
	WithdrawCompletedInt = 2
	WithdrawAbnormalInt  = 3
)

const (
	WithdrawPendingString   = "受理中"
	WithdrawCompletedString = "提币成功"
	WithdrawAbnormalString  = "异常"
)

func WithdrawStatusDetail(statusInt int32) (statusString string) {
	switch statusInt {
	case WithdrawPendingInt:
		statusString = WithdrawPendingString
	case WithdrawCompletedInt:
		statusString = WithdrawCompletedString
	case WithdrawAbnormalInt:
		statusString = WithdrawAbnormalString
	}
	return
}

// 提币记录
type WithdrawInfo struct {
	ID            uint64    `json:"id"`            //ID编号
	Sn           string    	`json:"sn"`            // 交易流水号
	Txid         string    	`json:"txid"`          // 交易哈希
	Amount       float64   	`json:"amount"`        // 提币金额
	UserName      string    `json:"user_name"`     // 用户名
	Fee          float64   	`json:"fee"`           // 手续费
	Gas          float64   	`json:"gas"`           // 矿工费
	Address      string    	`json:"address"`       // 目标地址
	Status       int32      `json:"status"`        // 状态码
	StatusString string    	`json:"status_string"` // 状态详情
	DoneAt       time.Time 	`json:"done_at"`       // 完成时间
	CreatedAt     time.Time `json:"created_at"`    //创建时间
	Currency     string     `json:"currency"`      // 代币类型
	Note         string    	`json:"note"`          // 提币备注
}

// 币种对应的提币地址
type WithdrawAddress struct {
	AccountID   uint64  `json:"account_id"`   // 账户ID
	Address     string 	`json:"address"`      // 地址
	AddressDesc string 	`json:"address_desc"` // 地址说明
	Currency    uint64  `json:"currency"`     // 代币ID
	Code        string 	`json:"code"`         // 代币代码
	Icon        string 	`json:"icon"`         // 代币图标
}

// WithdrawFromRedis 从redis中取出提币记录
func WithdrawFromRedis(tx string) (Withdraw, error) {
	withdrawKey := fmt.Sprintf("%s%s", WithdrawTx, tx)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	var withdraw Withdraw

	withdrawRecord, redisErr := redis.Bytes(RedisConn.Do("GET", withdrawKey))
	if redisErr != nil {
		// 获取失败，再尝试从数据库中查询
		if err := DB.Where("txid = ? and status = ?", tx, WithdrawPendingInt).First(&withdraw).Error; err != nil {
			return Withdraw{}, err
		}

		return withdraw, nil
	}

	bytesErr := json.Unmarshal(withdrawRecord, &withdraw)
	if bytesErr != nil {
		return Withdraw{}, bytesErr
	}
	return withdraw, nil
}

// WithdrawToRedis 将提币记录存到redis
func WithdrawToRedis(withdraw Withdraw) error {
	withdrawBytes, bytesErr := json.Marshal(withdraw)
	if bytesErr != nil {
		return bytesErr
	}
	withdrawKey := fmt.Sprintf("%s%s", WithdrawTx, withdraw.Txid)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	if _, redisErr := RedisConn.Do("SET", withdrawKey, withdrawBytes); redisErr != nil {
		return redisErr
	}
	return nil
}

// WithdrawDeleteRedis 将提币记录从redis中删除
func WithdrawDeleteRedis(tx string) error {
	withdrawKey := fmt.Sprintf("%s%s", WithdrawTx, tx)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	if _, redisErr := RedisConn.Do("DEL", withdrawKey); redisErr != nil {
		return redisErr
	}

	return nil
}

// 提币信息
type WithdrawExportInfo struct {
	ID           uint64     `json:"id"`            // 交易ID
	Sn           string    	`json:"sn"`            // 交易流水号
	Txid         string    	`json:"txid"`          // 交易哈希
	Amount       float64   	`json:"amount"`        // 提币金额
	Fee          float64   	`json:"fee"`           // 手续费
	Gas          float64   	`json:"gas"`           // 矿工费
	Address      string    	`json:"address"`       // 目标地址
	Status       int32      `json:"status"`        // 状态码
	StatusString string    	`json:"status_string"` // 状态详情
	DoneAt       string 	`json:"done_at"`       // 完成时间
	CurrencyCode string    	`json:"currency_code"` // 代币代码
	Note         string    	`json:"note"`          // 提币备注
	UserName     string    	`json:"user_name"`     //提币用户
	CreatedAt    string 	`json:"created_at"`    //创建时间
}
