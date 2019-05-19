package service

import (
	"fmt"
	"sync"
	"time"

	"galaxyotc/common/log"
	"galaxyotc/common/data"
	"galaxyotc/common/model"
)

type TokenService struct {
	tokenInfo map[string]*TokenItem
	tokenLock *sync.RWMutex
}

func newTokenService() *TokenService {
	service := &TokenService{
		tokenInfo: make(map[string]*TokenItem),
		tokenLock: new(sync.RWMutex),
	}
	service.loadTokens()
	return service
}

//缓存
type TokenItem struct {
	LoginStatus int32
	DeviceToken string
	Platform    int32
	PushType    int32
}

// load 友盟的token数据
func (s *TokenService) loadTokens() {
	s.tokenLock.Lock()
	defer s.tokenLock.Unlock()

	ts := []*model.PushToken{}

	if err := model.DB.Find(&ts).Error; err != nil {
		log.Error(err.Error())
		return
	}

	for _, t := range ts {
		item := &TokenItem{}
		item.DeviceToken = t.DeviceToken
		item.LoginStatus = t.LoginStatus
		item.Platform = t.Platform
		item.PushType = t.PushType
		s.tokenInfo[fmt.Sprintf("%d-%d", t.AppId, t.UserId)] = item
	}
}

func (s *TokenService) Save(req *data.PushDeviceInfoReq, userId uint64, platform int32) (bool, string) {
	var count int
	if err := model.DB.Table("push_token").Where(model.PushToken{AppId: req.AppId, UserId: userId}).Count(&count).Error; err != nil {
		log.Errorf("查友盟token是否重复失败,err: %s", err.Error())
		return false, "内部错误"
	}

	type UserInfo struct {
		UserId uint `json:"user_id"`
	}

	var pushToken model.PushToken
	if err := model.DB.Find(&pushToken, model.PushToken{DeviceToken: req.DeviceToken, LoginStatus: 1}).Error; err != nil {
		if err.Error() != "record not found" {
			log.Errorf("查友盟token是否重复失败,err: %s", err.Error())
			return false, "记录不存在"
		}
	} else {
		if uint64(pushToken.UserId) != userId {
			go s.Logout(userId, req.AppId)
		}
	}

	if count > 0 { //更新
		if err := model.DB.Table("push_token").Where(model.PushToken{UserId: userId, AppId: req.AppId}).Updates(model.PushToken{DeviceToken: req.DeviceToken, Platform: platform, LoginStatus: 1, UpdateTime: time.Now().Unix(), PushType: req.PushType}).Error; err != nil {
			return false, "更新token失败"
		}
	} else { //插入
		token := &model.PushToken{
			UserId: userId,
			AppId: req.AppId,
			DeviceToken: req.DeviceToken,
			Platform: platform,
			LoginStatus: model.ELOGIN_STATUS_LOGIN,
			UpdateTime: time.Now().Unix(),
			PushType: req.PushType,
		}

		if err := model.DB.Create(&token).Error; err != nil {
			return false, "添加token失败"
		}
	}

	//添加or更新缓存
	item := &TokenItem{}
	item.DeviceToken = req.DeviceToken
	item.LoginStatus = model.ELOGIN_STATUS_LOGIN
	item.Platform = platform
	item.PushType = req.PushType
	s.tokenInfo[fmt.Sprintf("%d-%d", req.AppId, userId)] = item

	return true, "success"
}

func (s *TokenService) SaveCache(appId int8, deviceToken string, pushType int32, userId uint64, platform int32) {
	//添加or更新缓存
	item := &TokenItem{}
	item.DeviceToken = deviceToken
	item.LoginStatus = model.ELOGIN_STATUS_LOGIN
	item.Platform = platform
	item.PushType = pushType
	s.tokenInfo[fmt.Sprintf("%d-%d", appId, userId)] = item
}

// 退出登录
func (s *TokenService) Logout(userId uint64, appId int32) bool {
	value, ok := s.tokenInfo[fmt.Sprintf("%d-%d", appId, userId)]
	if !ok {
		return false
	}

	if err := model.DB.Table("push_token").Where(model.PushToken{UserId: userId, AppId: appId}).Update(model.PushToken{LoginStatus: 0}).Error; err != nil {
		return false
	}

	//更新缓存
	value.LoginStatus = model.ELOGIN_STATUS_NOLOGIN

	s.tokenInfo[fmt.Sprintf("%d-%d", appId, userId)] = value

	return true
}

// 得到缓存数据
func (s *TokenService) GetTokenCache(key string) *TokenItem {
	value, ok := s.tokenInfo[key]
	if !ok {
		return nil
	}

	return value
}
