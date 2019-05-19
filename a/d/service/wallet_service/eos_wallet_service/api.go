package ethereum_wallet_service

import (
	"galaxyotc/common/log"
	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/eos_wallet"
	"sync"
	"github.com/nats-rpc/nrpc"
	"time"
)

var client *Client

type Client struct {
	*pb.EosWalletServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			EosWalletServiceClient: pb.NewEosWalletServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 获取充值地址
func (c *Client) Deposit(purpose wi.KeyPurpose) (string, error) {
	resp, err := c.EosWalletServiceClient.Deposit(pb.DepositReq{int32(purpose)})
	if err != nil {
		log.Errorf("api-Deposit-error: %s", err.Error())
		return "", err
	}
	return resp.Address, nil
}

// 提现申请：发送EOS给接收者
func (c *Client) EosWithdraw(to, value string) (string, error) {
	resp, err := c.EosWalletServiceClient.EosWithdraw(pb.EosWithdrawReq{to, value})
	if err != nil {
		log.Errorf("api-EosWithdraw-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}