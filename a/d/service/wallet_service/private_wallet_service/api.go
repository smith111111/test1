package private_wallet_service

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/wallet/private_wallet"
	"sync"
	"github.com/nats-rpc/nrpc"
	"time"
)

const (
	ToBuyer int32 = 0
	ToSeller int32 = 1
)

var client *Client

type Client struct {
	*pb.PrivateWalletServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			PrivateWalletServiceClient: pb.NewPrivateWalletServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 区块当前高度
func (c *Client) ChainTip() (uint32, string, error) {
	resp, err := c.PrivateWalletServiceClient.ChainTip(pb.ChainTipReq{})
	if err != nil {
		log.Errorf("api-ChainTip-error: %s", err.Error())
		return 0, "", err
	}
	return resp.Height, resp.Hash, nil
}

// 给指定地址转账
func (c *Client) Transfer(to, value string) (string, error) {
	resp, err := c.PrivateWalletServiceClient.Transfer(pb.TransferReq{to, value})
	if err != nil {
		log.Errorf("api-Transfer-error: %s", err.Error())
		return "", err
	}
	return resp.Hash, nil
}

// 获取一个新地址
func (c *Client) NewAddress(purpose int32) (string, error) {
	resp, err := c.PrivateWalletServiceClient.NewAddress(pb.NewAddressReq{purpose})
	if err != nil {
		log.Errorf("api-NewAddress-error: %s", err.Error())
		return "", err
	}
	return resp.Address, nil
}

// 添加一个代币担保交易：即锁币
func (c *Client) AddTokenTransaction(amount, sn string, threshold int32, timeout int64, buyerAddress, sellerAddress, tokenAddress string) (string, error) {
	resp, err := c.PrivateWalletServiceClient.AddTokenTransaction(pb.TransactionReq{Amount: amount, Sn: sn, Threshold: threshold, Timeout: timeout, BuyerAddress: buyerAddress, SellerAddress: sellerAddress, TokenAddress: tokenAddress})
	if err != nil {
		log.Errorf("api-AddTokenTransaction-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}

// 构造参数并释放一个代币担保交易:即放币
func (c *Client) ExecuteTransaction(amount, sn string, threshold int32, timeout int64, buyerAddress, sellerAddress, tokenAddress string, to int32) (string, error) {
	resp, err := c.PrivateWalletServiceClient.ExecuteTransaction(pb.TransactionReq{Amount: amount, Sn: sn, Threshold: threshold, Timeout: timeout, BuyerAddress: buyerAddress, SellerAddress: sellerAddress, TokenAddress: tokenAddress, To: to})
	if err != nil {
		log.Errorf("api-ExecuteTransaction-error: %s", err.Error())
		return "", err
	}
	return resp.Txid, nil
}

// 部署代币合约
func (c *Client) DeployToken(name, symbol string, decimals uint32) (string, error) {
	resp, err := c.PrivateWalletServiceClient.DeployToken(pb.DeployTokenReq{name, symbol, decimals})
	if err != nil {
		log.Errorf("api-DeployToken-error: %s", err.Error())
		return "", err
	}
	return resp.Address, nil
}

// 获取代币的余额
func (c *Client) GetTokenBalance(tokenAddress, whoAddress string) (string, error) {
	resp, err := c.PrivateWalletServiceClient.GetTokenBalance(pb.GetTokenBalanceReq{tokenAddress, whoAddress})
	if err != nil {
		log.Errorf("api-GetTokenBalance-error: %s", err.Error())
		return "", err
	}
	return resp.Balance, nil
}

// 挖矿代币
func (c *Client) MintToken(tokenAddress, whoAddress, amount string) (error) {
	_, err := c.PrivateWalletServiceClient.MintToken(pb.TokenReq{tokenAddress, whoAddress, amount})
	if err != nil {
		log.Errorf("api-MintToken-error: %s", err.Error())
		return err
	}
	return nil
}

// 燃烧代币
func (c *Client) BurnToken(tokenAddress, whoAddress, amount string) (error) {
	_, err := c.PrivateWalletServiceClient.BurnToken(pb.TokenReq{tokenAddress, whoAddress, amount})
	if err != nil {
		log.Errorf("api-BurnToken-error: %s", err.Error())
		return err
	}
	return nil
}