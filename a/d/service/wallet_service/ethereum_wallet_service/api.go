package ethereum_wallet_service

import (
	"galaxyotc/common/log"
	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/ethereum_wallet"
	"sync"
	"github.com/nats-rpc/nrpc"
	"time"
)

var client *Client

type Client struct {
	*pb.EthereumWalletServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			EthereumWalletServiceClient: pb.NewEthereumWalletServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 获取充值地址
func (c *Client) Deposit(purpose wi.KeyPurpose) (string, error) {
	resp, err := c.EthereumWalletServiceClient.Deposit(pb.DepositReq{int32(purpose)})
	if err != nil {
		log.Errorf("api-Deposit-error: %s", err.Error())
		return "", err
	}
	return resp.Address, nil
}

// 提现申请：发送ETH给接收者
func (c *Client) EtherWithdraw(to, value string) (string, error) {
	resp, err := c.EthereumWalletServiceClient.EtherWithdraw(pb.EtherWithdrawReq{to, value})
	if err != nil {
		log.Errorf("api-EtherWithdraw-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}

// 提现申请：发送ERC20代币给接收者
func (c *Client) TokenWithdraw(tokenAddr, to, value string) (string, error) {
	resp, err := c.EthereumWalletServiceClient.TokenWithdraw(pb.TokenWithdrawReq{tokenAddr, to, value})
	if err != nil {
		log.Errorf("api-TokenWithdraw-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}