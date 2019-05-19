package exchangerate_service

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/exchange_rate"
	"github.com/nats-rpc/nrpc"
	"sync"
	"time"
)

var client *Client

type Client struct {
	*pb.ExchangeRateServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			ExchangeRateServiceClient: pb.NewExchangeRateServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 获取所有汇率
func (c *Client) GetAllRates(cache bool) (map[string]float64, error) {
	resp, err := c.ExchangeRateServiceClient.GetAllRates(pb.GetAllRatesReq{Cache: cache})
	if err != nil {
		log.Errorf("api-GetAllRates-error: %s", err.Error())
		return nil, err
	}
	return resp.AllRates, nil
}

// 获取指定币种的汇率（即，1个BTC的价格）
func (c *Client) GetExchangeRate(code string) (float64, error) {
	resp, err := c.ExchangeRateServiceClient.GetExchangeRate(pb.GetExchangeRateReq{Code: code})
	if err != nil {
		log.Errorf("api-GetExchangeRate-error: %s", err.Error())
		return 0, err
	}
	return resp.Rate, nil
}

// 获取最新指定币种的汇率（即，1个BTC的价格），并更新缓存
func (c *Client) GetLatestRate(code string) (float64, error) {
	resp, err := c.ExchangeRateServiceClient.GetLatestRate(pb.GetExchangeRateReq{Code: code})
	if err != nil {
		log.Errorf("api-GetLatestRate-error: %s", err.Error())
		return 0, err
	}
	return resp.Rate, nil
}