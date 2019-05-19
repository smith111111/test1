package multi_wallet_service

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/wallet/multi_wallet"
	bwi "galaxyotc/btc-wallet-interface"
	"github.com/nats-rpc/nrpc"
	"sync"
	"time"
)

var client *Client

type Client struct {
	*pb.MultiWalletServiceClient
}

type OmniTransaction struct {
	Amount 			string
	Fee				string
	Txid			string
	BlockTime		time.Time
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			MultiWalletServiceClient: pb.NewMultiWalletServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 区块当前高度
func (c *Client) ChainTip(code, propertyId string) (uint32, string, error) {
	resp, err := c.MultiWalletServiceClient.ChainTip(pb.ChainTipReq{code, propertyId})
	if err != nil {
		log.Errorf("api-ChainTip-error: %s", err.Error())
		return 0, "", err
	}
	return resp.Height, resp.Hash, nil
}

// 获取充值地址
func (c *Client) Deposit(code, propertyId string, purpose bwi.KeyPurpose) (string, error) {
	resp, err := c.MultiWalletServiceClient.Deposit(pb.DepositReq{code, propertyId, int32(purpose)})
	if err != nil {
		log.Errorf("api-Deposit-error: %s", err.Error())
		return "", err
	}
	return resp.Address, nil
}

// 提币
func (c *Client) Withdraw(code, propertyId, address string, amount int64, feeLevel int32) (string, error) {
	resp, err := c.MultiWalletServiceClient.Withdraw(pb.WithdrawReq{code, propertyId, address, amount, feeLevel})
	if err != nil {
		log.Errorf("api-Withdraw-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}

// 获取Omni交易信息
func (c *Client) GetOmniTransaction(code, txid string) (*OmniTransaction, error) {
	resp, err := c.MultiWalletServiceClient.GetOmniTransaction(pb.GetOmniTransactionReq{code, txid})
	if err != nil {
		log.Errorf("api-GetOmniTransaction-error: %s", err.Error())
		return nil, err
	}

	transaction := &OmniTransaction{
		resp.Amount,
		resp.Fee,
		resp.Txid,
		time.Unix(resp.BlockTime, 0),
	}
	return transaction, nil
}