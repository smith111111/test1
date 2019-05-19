package user

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

// TestGetUserResult  测试获取用户信息
func TestGetUserResult(t *testing.T) {
	c := newClient()

	result, err := c.GetUser(1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result is : %+v", result)
}

// TestIsExistResult  检查用户是否是否存在
func TestIsExistResult(t *testing.T) {
	c := newClient()

	result, err := c.IsExist("galaxy@otc.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result is : %t", result)
}
