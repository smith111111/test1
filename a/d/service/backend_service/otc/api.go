package user

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/backend/otc"
	"github.com/nats-rpc/nrpc"
	"sync"
	"time"
)

var client *Client

type Client struct {
	*pb.OTCServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			OTCServiceClient: pb.NewOTCServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 订单超时回调
func (c *Client) OrderTimeoutCallback(callbackType int32, sn string, scene int32) (bool, error) {
	resp, err := c.OTCServiceClient.OrderTimeoutCallback(pb.OrderTimeoutCallbackReq{callbackType, sn, scene})
	if err != nil {
		log.Errorf("api-OrderTimeoutCallback-error: %s", err.Error())
		return pb.BoolResp{}.Ok, err
	}
	return resp.Ok, nil
}