package user

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/backend/account"
	"sync"
	"github.com/nats-rpc/nrpc"
	"time"
)

var client *Client

type Client struct {
	*pb.AccountServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			AccountServiceClient: pb.NewAccountServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// EthereumWallet监听回调
func (c *Client) EthereumCallback(transaction []byte) (bool, error) {
	resp, err := c.AccountServiceClient.EthereumCallback(pb.CallbackReq{transaction})
	if err != nil {
		log.Errorf("api-EthereumCallback-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}

// EosWallet监听回调
func (c *Client) EosCallback(transaction []byte) (bool, error) {
	resp, err := c.AccountServiceClient.EosCallback(pb.CallbackReq{transaction})
	if err != nil {
		log.Errorf("api-EosCallback-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}

// MultiWallet监听回调
func (c *Client) MultiCallback(transaction []byte) (bool, error) {
	resp, err := c.AccountServiceClient.MultiCallback(pb.CallbackReq{transaction})
	if err != nil {
		log.Errorf("api-MultiCallback-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}

// PrivateWallet错误处理回调
func (c *Client) PrivateErrorCallback(transaction []byte) (bool, error) {
	resp, err := c.AccountServiceClient.PrivateErrorCallback(pb.CallbackReq{transaction})
	if err != nil {
		log.Errorf("api-PrivateErrorCallback-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}

// 账户转账交易
func (c *Client) AccountTransfer(amount, sn, receiverAddress, senderAddress, tokenAddress string) (bool, error) {
	resp, err := c.AccountServiceClient.AccountTransfer(pb.AccountTransferReq{Amount: amount, Sn: sn, ReceiverAddress: receiverAddress, SenderAddress: senderAddress, TokenAddress: tokenAddress})
	if err != nil {
		log.Errorf("api-AccountTransfer-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}