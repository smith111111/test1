package captcha

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/task"
	"sync"
	"github.com/nats-rpc/nrpc"
	"time"
)

type timeoutType int
type cancelScene int

// 超时回调类型
const (
	// 超时取消
	OrderCancel timeoutType = 1
	// 超时放币
	OrderRelease timeoutType = 2
	// 超时完成
	OrderCompleted timeoutType = 3
)

// 订单取消场景
const (
	// 等待接单
	WaitingApproved cancelScene = 1
	// 等待付款
	WaitingPay cancelScene = 2
)

var client *Client

type Client struct {
	*pb.TaskServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			TaskServiceClient: pb.NewTaskServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 订单超时回调
func (c *Client) OrderTimeout(duration int64, timeoutType timeoutType, orderSn string, scenes ...cancelScene) (bool, error) {
	var scene int32
	if len(scenes) <= 0 {
		scene = 0
	} else {
		scene = int32(scenes[0])
	}
	resp, err := c.TaskServiceClient.OrderTimeout(pb.OrderTimeoutReq{duration, int32(timeoutType), orderSn, scene})
	if err != nil {
		log.Errorf("api-OrderTimeout-error: %s", err.Error())
		return false, err
	}
	return resp.Ok, nil
}

// 账户转账交易
func (c *Client) AccountTransfer(amount, sn, receiverAddress, senderAddress, tokenAddress string) (bool, error) {
	resp, err := c.TaskServiceClient.AccountTransfer(pb.AccountTransferReq{Amount: amount, Sn: sn, ReceiverAddress: receiverAddress, SenderAddress: senderAddress, TokenAddress: tokenAddress})
	if err != nil {
		log.Errorf("api-AccountTransfer-error: %s", err.Error())
		return false, err
	}
	return resp.Ok, nil
}