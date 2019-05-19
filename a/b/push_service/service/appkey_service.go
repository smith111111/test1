package service

import (
	"fmt"
	"sync"

	"galaxyotc/common/log"

	"galaxyotc/common/model"
)

type AppkeyService struct {
	appkeyInfo map[string]*AppkeyItem
	appkeyLock *sync.RWMutex
}

func newAppkeyService() *AppkeyService {
	service := &AppkeyService{
		appkeyInfo: make(map[string]*AppkeyItem),
		appkeyLock: new(sync.RWMutex),
	}
	service.loadAppkey()
	return service
}

type AppkeyItem struct {
	Appkey      string
	Secret      string
	PackageName string
}

type AppkeyAddReq struct {
	Appid       int    `json:"appid"`
	Platform    int    `json:"platform"`
	PushType    int    `json:"push_type"`
	Appkey      string `json:"appkey"`
	Secret      string `json:"secret"`
	PackageName string `json:"package_name"`
}

func (s *AppkeyService) loadAppkey() {
	s.appkeyLock.Lock()
	defer s.appkeyLock.Unlock()

	aks := []*model.PushAppKey{}

	if err := model.DB.Find(&aks).Error; err != nil {
		log.Error(err.Error())
		return
	}

	for _, ak := range aks {
		item := &AppkeyItem{}
		item.Appkey = ak.AppKey
		item.Secret = ak.Secret
		item.PackageName = ak.PackageName
		s.appkeyInfo[fmt.Sprintf("%d&%d&%d", ak.AppId, ak.Platform, ak.PushType)] = item
	}

	//log.Infof("load appkey data,len is :%d", len(p.appkeyInfo))
}

//得到appid和plat得到appkey和秘钥
func (s *AppkeyService) GetAppkeyItem(appid int32, plat int32, pushType int32) *AppkeyItem {
	key := fmt.Sprintf("%d&%d&%d", appid, plat, pushType)
	value, ok := s.appkeyInfo[key]
	if !ok {
		return nil
	}

	return value
}
