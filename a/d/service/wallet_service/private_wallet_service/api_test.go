package private_wallet_service

import (
	"testing"
	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func newClient() *Client {
	// 获取nats连接
	natsUrl := "nats://192.168.0.231:4222"
	serviceName := "user_service_test"
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	return NewClient(nc)
}

// TestChainTip 获取区块当前高度
func TestChainTip(t *testing.T) {
	c := newClient()

	height, hash, err := c.ChainTip()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("height is : %d, hash is : %s", height, hash)
}

// TestDeployToken 部署代币合约
func TestDeployToken(t *testing.T) {
	c := newClient()

	result, err := c.DeployToken("EOS", "EOS", 4)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result is : %s", result)
}

// TestGetTokenBalance 获取代币的余额
func TestGetTokenBalance(t *testing.T) {
	c := newClient()

	result, err := c.GetTokenBalance("0xbB95CD8dd415d158120cfa14402d83a45aAF1777", "0x75cdFc99a0a3Ecb29800c5aA02D007156e27ee29")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result is : %s", result)

	balance, _ := utils.ToDecimal(result, 18).Float64()

	t.Logf("balance is : %f", balance)
}

// TestTransfer 给指定地址转账
func TestTransfer(t *testing.T) {
	c := newClient()

	result, err := c.Transfer("0x8A565d31C3002BB16b5caE6276cd181B8523C14E", "100000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result is : %s", result)
}