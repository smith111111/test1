package user

import (
	"galaxyotc/common/log"
	pb "galaxyotc/common/proto/backend/user"
	"github.com/nats-rpc/nrpc"
	"sync"
	"time"
)

var client *Client

type Client struct {
	*pb.UserServiceClient
}

type UserInfo struct {
	Id              uint64
	Name            string
	AreaCode        string
	Mobile          string
	Email           string
	AvatarUrl       string
	Status          int32
	InternalAddress string
	ParentId        uint64
	UserType        int32
	DiscountRate    float64
	IsRealName      bool
	ReferralCode    string
	TradingMethods  string
}

// 创建新连接
func NewClient(nc nrpc.NatsConn) *Client {
	var once sync.Once
	once.Do(func() {
		client = &Client{
			UserServiceClient: pb.NewUserServiceClient(nc),
		}
		client.Timeout = 60 * time.Second
	})
	return client
}

// 获取用户信息
func (c *Client) GetUser(id uint64) (*UserInfo, error) {
	resp, err := c.UserServiceClient.GetUser(pb.GetUserReq{UserId: id})
	if err != nil {
		log.Errorf("api-GetUser-error: %s", err.Error())
		return nil, err
	}

	user := &UserInfo{
		Id: resp.Id,
		Name: resp.Name,
		AreaCode: resp.AreaCode,
		Mobile: resp.Mobile,
		Email: resp.Email,
		AvatarUrl: resp.AvatarUrl,
		Status: resp.Status,
		InternalAddress: resp.InternalAddress,
		ParentId: resp.ParentId,
		UserType: resp.UserType,
		DiscountRate: resp.DiscountRate,
		IsRealName: resp.IsRealName,
		ReferralCode: resp.ReferralCode,
		TradingMethods: resp.TradingMethods,
	}
	return user, nil
}

// 获取用户信息
func (c *Client) GetUserByInternalAddress(internalAddress string) (*UserInfo, error) {
	resp, err := c.UserServiceClient.GetUserByInternalAddress(pb.GetUserByInternalAddressReq{InternalAddress: internalAddress})
	if err != nil {
		log.Errorf("api-GetUserByInternalAddress-error: %s", err.Error())
		return nil, err
	}

	user := &UserInfo{
		Id: resp.Id,
		Name: resp.Name,
		AreaCode: resp.AreaCode,
		Mobile: resp.Mobile,
		Email: resp.Email,
		AvatarUrl: resp.AvatarUrl,
		Status: resp.Status,
		InternalAddress: resp.InternalAddress,
		ParentId: resp.ParentId,
		UserType: resp.UserType,
		DiscountRate: resp.DiscountRate,
		IsRealName: resp.IsRealName,
		ReferralCode: resp.ReferralCode,
		TradingMethods: resp.TradingMethods,
	}
	return user, nil
}

// 检查用户是否是否存在
func (c *Client) IsExist(input string) (bool, error) {
	resp, err := c.UserServiceClient.IsExist(pb.IsExistReq{Input: input})
	if err != nil {
		log.Errorf("api-IsExist-error: %s", err.Error())
		return false, err
	}
	return resp.Exist, nil
}