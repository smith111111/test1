package model

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"errors"
	"time"
)

// 法币信息
type FiatCurrency struct {
	Code  					string 		`gorm:"size:20; unique_index;not null" json:"code"`		// 法币代码
	Name 					string		`gorm:"size:20; unique_index;not null" json:"name"` 	// 法币名称
	CreatedAt           	time.Time 	`json:"created_at"`                            			// 创建时间
	UpdatedAt           	time.Time 	`json:"updated_at"`                            			// 更新时间
	Symbol					string 		`gorm:"size:5" json:"symbol"`							// 法币符号
	Icon 					string 		`json:"icon"`											// 图标
	Status					int32		`gorm:"default:0" json:"status"`						// 状态
	Precision				uint32		`json:"precision"`										// 小数点位数
}

// FiatCurrencyDeleteRedis 将对应的法币信息用redis中删除
func FiatCurrencyDeleteRedis(fiat FiatCurrency) error {
	fiatCurrencyByCode := fmt.Sprintf("%s%s", FiatCurrencyByCode, fiat.Code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	// 根据业务需求的不同，可以根据代币ID或代币代码获取代币信息
	if _, redisErr := RedisConn.Do("DEL", fiatCurrencyByCode); redisErr != nil {
		fmt.Println("redis set failed: ", redisErr.Error())
		return errors.New("error")
	}
	return nil
}

// FiatCurrencyFromAndToRedis 根据法币ID或者法币代码从redis中取出法币信息
func FiatCurrencyFromAndToRedis(code string) (FiatCurrency, error) {
	fiatCurrencyKey := fmt.Sprintf("%s%s", FiatCurrencyByCode, code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	fiatBytes, err := redis.Bytes(RedisConn.Do("GET", fiatCurrencyKey))
	if err != nil {
		// Redis中获取不到数据时，再从mysql中查询
		var fiatCurrency FiatCurrency
		if err := DB.First(&fiatCurrency, FiatCurrency{Code: code, Status: CurrencyNormalInt}).Error; err != nil {
			return fiatCurrency, err
		}
		// 查询之后在存入Redis中
		FiatCurrencyToRedis(fiatCurrency)
		return fiatCurrency, nil
	}

	var fiatCurrency FiatCurrency
	bytesErr := json.Unmarshal(fiatBytes, &fiatCurrency)
	if bytesErr != nil {
		return fiatCurrency, errors.New("获取法币信息失败")
	}
	return fiatCurrency, nil
}

// FiatCurrencyToRedis 将法币信息存到redis
func FiatCurrencyToRedis(fiat FiatCurrency) error {
	fiatBytes, err := json.Marshal(fiat)
	if err != nil {
		fmt.Println(err)
		return errors.New("error")
	}

	// 根据代币代码存储
	FiatCurrencyByCodeKey := fmt.Sprintf("%s%s", FiatCurrencyByCode, fiat.Code)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	// 根据业务需求的不同，可以根据代币ID或代币代码获取代币信息
	if _, redisErr := RedisConn.Do("SET", FiatCurrencyByCodeKey, fiatBytes); redisErr != nil {
		fmt.Println("redis set failed: ", redisErr.Error())
		return errors.New("error")
	}
	return nil
}