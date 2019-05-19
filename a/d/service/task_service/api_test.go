package captcha

import (
	"testing"
	commonService "galaxyotc/common/service"
)

func newClient() *Client {
	// 获取nats连接
	natsUrl := "nats://192.168.0.231:4222"
	serviceName := "task_service_test"
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	return NewClient(nc)
}

// TestAccountTransfer  测试转账交易
func TestAccountTransfer(t *testing.T) {
	c := newClient()

	result, err := c.AccountTransfer("1000000000000000000", "bgviken01cnoga9ithh0",  "0x3211347DD711a036E564a262fCcE9800075fa214", "0x8A565d31C3002BB16b5caE6276cd181B8523C14E","0xbB95CD8dd415d158120cfa14402d83a45aAF1777")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result is : %+v", result)
}