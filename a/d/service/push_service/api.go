package push_service

import (
	"sync"
	"errors"

	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/push"
	"github.com/nats-rpc/nrpc"
	"github.com/gin-gonic/gin/json"
	"time"
)

var client *Client

type Client struct {
	*pb.PushP2PClient
	*pb.PushP2MClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			PushP2PClient: pb.NewPushP2PClient(nc),
			PushP2MClient: pb.NewPushP2MClient(nc),
		}
		client.PushP2PClient.Timeout = 60 * time.Second
		client.PushP2MClient.Timeout = 60 * time.Second
	})
	return client
}

// 列播(含单播)
func (c *Client) SendMsg(displayType string, receivers string, appId int32, title string, text string, custom map[string]interface{}, loginStatus int32) error {
	buf, err := json.Marshal(custom)
	if err != nil {
		log.Errorf("api-SendMsg-error: %s", err.Error())
		return err
	}

	expireTime := int32(time.Now().AddDate(0, -3, 0).Unix())

	resp, err := c.PushP2PClient.SendMsg(pb.SendMsgReq{DisplayType: displayType, Receivers: receivers, AppId: appId, Title: title, Text: text, Custom: buf, ExpireTime: expireTime, LoginStatus: loginStatus})
	if err != nil {
		log.Errorf("api-SendMsg-error: %s", err.Error())
		return err
	}

	if resp.ErrNo != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}

// 用户登出
func (c *Client) Logout(userId uint64, appId int32) error {
	resp, err := c.PushP2PClient.Logout(pb.LogoutReq{UserId: userId, AppId: appId})
	if err != nil {
		log.Errorf("api-Logout-error: %s", err.Error())
		return err
	}

	if resp.ErrNo != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}

// 同步设备相关信息
func (c *Client) BroadcastDeviceInfo(appId int32, deviceToken string, pushType int32, userId uint64, platform int32) error {
	resp, err := c.PushP2MClient.SyncDeviceInfo(pb.DeviceInfoReq{AppId: appId, DeviceToken: deviceToken, PushType: pushType, Uid: userId, Platform: platform})
	if err != nil {
		log.Errorf("api-Logout-error: %s", err.Error())
		return err
	}

	if resp.ErrNo != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}