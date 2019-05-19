package front_service

import (
	"sync"
	"errors"
	"time"

	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/front"
	"github.com/nats-rpc/nrpc"
	"github.com/gin-gonic/gin/json"
)

var client *Client

type Client struct {
	*pb.FrontP2MClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			FrontP2MClient: pb.NewFrontP2MClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

//接收任务
func (c *Client) RecvPush(uid uint64, msgId int64, title string, text string, custom map[string]interface{}) error {
	buf, err := json.Marshal(custom)
	if err != nil {
		log.Errorf("api-RecvPush-error: %s", err.Error())
		return err
	}

	resp, err := c.FrontP2MClient.RecvPush(pb.RecvPushReq{Uid: uid, MsgId: msgId, Title: title, Text: text, Custom: buf})
	if err != nil {
		log.Errorf("api-RecvPush-error: %s", err.Error())
		return err
	}

	if resp.ErrNo != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}