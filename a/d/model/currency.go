package model

import (
	"fmt"
	"errors"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"time"
)

// 代币信息
type Currency struct {
	Code                string    `gorm:"primary_key; size:20;" json:"code"`     // 代币代码
	Name                string    `gorm:"size:20" json:"name"`                   // 代币名称
	CreatedAt           time.Time `json:"created_at"`                            // 创建时间
	UpdatedAt           time.Time `json:"updated_at"`                            // 更新时间
	Symbol              string    `gorm:"size:5" json:"symbol"`                  // 代币符号
	Family              string    `gorm:"index" json:"family"`                   // 代币系列
	Icon                string    `json:"icon"`                                  // 图标
	Status              int32     `gorm:"default:0" json:"status"`               // 状态
	PropertyId          string    `gorm:"size:20; default:0" json:"property_id"` // 资产ID
	Precision           uint32    `gorm:"default:18" json:"precision"`           // 小数点位数
	LowFee              uint32    `gorm:"default:140" json:"low_fee"`            // 矿工费
	MediumFee           uint32    `gorm:"default:160" json:"medium_fee"`
	HighFee             uint32    `gorm:"default:180" json:"high_fee"`
	MaxFee              uint32    `gorm:"default:200" json:"max_fee"`
	PrivateTokenAddress string    `gorm:"not null" json:"private_token_address"` // 代币合约地址
	PublicTokenAddress  string    `json:"public_token_address"`                  // 外部合约地址
	FeeAPI              string    `json:"fee_api"`                               // 外部API，用于查询矿工费用。如果为nil，那么默认费用将被使用。 API响应格式：{"fastestFee": 40, "halfHourFee": 20, "hourFee": 10}
	ClientAPI           string    `json:"client_api"`                            // 信任的API，用于查询链上的余额和监听区块链事件
	WithdrawFee			float64   `json:"withdraw_fee"`							 // 提币手续费
	Sort 				int32	  `json:"sort"`									 // 排序顺序
}

const (
	CurrencyNormalInt  = 0
	CurrencyDisableInt = 1
)

const (
	CurrencyNormalString  = "正常"
	CurrencyDisableString = "禁用"
)

func (crypto *Currency) ChangeStatus(status int) error {
	if err := DB.Model(&crypto).Update("status", status).Error; err != nil {
		return err
	}

	// 根据修改状态选择是保存到redis还是从redis中删除
	switch status {
	case CurrencyNormalInt:
		_ = CurrencyToRedis(*crypto)
	case CurrencyDisableInt:
		_ = CurrencyDeleteRedis(*crypto)
	}

	return nil
}

func (crypto *Currency) StatusDetail() string {
	switch crypto.Status {
	case CurrencyNormalInt:
		return CurrencyNormalString
	case CurrencyDisableInt:
		return CurrencyDisableString
	}
	return ""
}

// CurrencyDeleteRedis 将对应的代币信息用redis中删除
func CurrencyDeleteRedis(currency Currency) error {
	CurrencyByCode := fmt.Sprintf("%s%s", CurrencyByCode, currency.Code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	if _, redisErr := RedisConn.Do("DEL", CurrencyByCode); redisErr != nil {
		fmt.Println("redis set failed: ", redisErr.Error())
		return errors.New("error")
	}
	return nil
}

// CurrencyFromRedis 根据代币ID或者代币代码从redis中取出代币信息
func CurrencyFromAndToRedis(code string) (Currency, error) {
	currencyKey := fmt.Sprintf("%s%s", CurrencyByCode, code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	currencyBytes, err := redis.Bytes(RedisConn.Do("GET", currencyKey))
	if err != nil {
		// Redis中获取不到数据时，再从mysql中查询
		var currency Currency
		if err := DB.Where("code=? and status=?", code, CurrencyNormalInt).First(&currency).Error; err != nil {
			return currency, err
		}

		// 查询之后在存入Redis中
		CurrencyToRedis(currency)

		return currency, nil
	}

	var currency Currency
	bytesErr := json.Unmarshal(currencyBytes, &currency)
	if bytesErr != nil {
		return currency, errors.New("获取代币信息失败")
	}
	return currency, nil
}

// CurrencyToRedis 将代币信息存到redis
func CurrencyToRedis(currency Currency) error {
	currencyBytes, err := json.Marshal(currency)
	if err != nil {
		fmt.Println(err)
		return errors.New("error")
	}

	CurrencyByCodeKey := fmt.Sprintf("%s%s", CurrencyByCode, currency.Code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	// 根据业务需求的不同，可以根据代币ID或代币代码获取代币信息
	if _, redisErr := RedisConn.Do("SET", CurrencyByCodeKey, currencyBytes); redisErr != nil {
		fmt.Println("redis set failed: ", redisErr.Error())
		return errors.New("error")
	}
	return nil
}