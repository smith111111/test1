package exchangerate_service

import (
	"testing"
	commonService "galaxyotc/common/service"
)

func newClient() *Client {
	// 获取nats连接
	natsUrl := "nats://192.168.0.231:4222"
	serviceName := "user_service_test"
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	return NewClient(nc)
}

// TestGetAllRates 获取所有汇率
func TestGetAllRates(t *testing.T) {
	c := newClient()

	result, err := c.GetAllRates(true)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result is : %v", result)
}

// TestGetExchangeRate 获取指定币种的汇率（即，1个BTC的价格）
func TestGetExchangeRate(t *testing.T) {
	c := newClient()

	result, err := c.GetExchangeRate("CNY")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result is : %f", result)
}

// TestGetLatestRate 获取最新指定币种的汇率（即，1个BTC的价格），并更新缓存
func TestGetLatestRate(t *testing.T) {
	c := newClient()

	result, err := c.GetLatestRate("CNY")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result is : %f", result)
}