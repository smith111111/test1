package im_service

import (
	"sync"
	"errors"

	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/im"
	"github.com/nats-rpc/nrpc"
	"github.com/gin-gonic/gin/json"
	"time"
)

var client *Client

type Client struct {
	*pb.IMServiceClient
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			IMServiceClient: pb.NewIMServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

func (c *Client) UserRegister(id uint64, name string, props map[string]interface{}, icon string, email string, birth string, mobile string, gender int8, ex map[string]interface{}) (string, error) {
	propsBuf, err := json.Marshal(props)
	if err != nil {
		log.Errorf("api-UserRegister-error: %s", err.Error())
		return "", err
	}

	if mobile != "" && len(mobile) > 11 {
		mobile = mobile[len(mobile)-11:]
	}

	exBuf, err := json.Marshal(ex)
	if err != nil {
		log.Errorf("api-UserRegister-error: %s", err.Error())
		return "", err
	}

	resp, err := c.IMServiceClient.UserRegister(pb.UserRegisterReq{Id: id, Name: name, Props: propsBuf, Icon: icon, Email: email, Birth: birth, Mobile: mobile, Gender: int32(gender), Ex: exBuf})
	if err != nil {
		log.Errorf("api-UserRegister-error: %s", err.Error())
		return "", err
	}

	return resp.Token, nil
}

// 批量发送自定义系统消息
func (c *Client) SendSysMsg(from uint64, to uint64, attach map[string]interface{}) error {
	buf, err := json.Marshal(attach)
	if err != nil {
		log.Errorf("api-SendSysMsg-error: %s", err.Error())
		return err
	}

	resp, err := c.IMServiceClient.SendSysMsg(pb.SendSysMsgReq{From: from, To: to, Attach: buf})
	if err != nil {
		log.Errorf("api-SendSysMsg-error: %s", err.Error())
		return err
	}

	if resp.ErrNo != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}