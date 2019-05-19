package auth

import "galaxyotc/common/data"

var (
	appkeyser *AppkeyService
)

//每个appkey的基本信息
type AppKeyItem struct {
	Desc      string //appkey的描述
	SecretKey string //私钥
}

func newAppKeyItem(desc, secrect string) *AppKeyItem {
	return &AppKeyItem{
		Desc:      desc,
		SecretKey: secrect,
	}
}

type AppkeyService struct {
	appkeyInfo map[string]*AppKeyItem
}

func (p *AppkeyService) initService() {
	p.appkeyInfo[data.APPKEY_OTC_ANDROID] = newAppKeyItem("otc android", data.SECRECT_OTC_ANDROID)
	p.appkeyInfo[data.APPKEY_OTC_IOS] = newAppKeyItem("otc ios", data.SECRECT_OTC_IOS)
	p.appkeyInfo[data.APPKEY_OTC_H5] = newAppKeyItem("otc h5", data.SECRECT_OTC_H5)
	p.appkeyInfo[data.APPKEY_OTC_BACKENDSERVER] = newAppKeyItem("otc backendserver", data.SECRECT_OTC_BACKENDSERVER)
}

func newappkeyService() *AppkeyService {
	service := &AppkeyService{}
	service.appkeyInfo = make(map[string]*AppKeyItem)
	service.initService()
	return service
}

func GetAuthAppKeyService() *AppkeyService {
	if appkeyser == nil {
		appkeyser = newappkeyService()
	}
	return appkeyser
}

//验证appkey是否合法
func (p *AppkeyService) ValidateAppKey(appkey string) bool {
	if _, ok := p.appkeyInfo[appkey]; ok {
		return true
	}
	return false
}

//根据appkey获取appkey的基本信息
func (p *AppkeyService) GetAppKeyItem(appkey string) (*AppKeyItem, bool) {
	item, ok := p.appkeyInfo[appkey]
	return item, ok
}

//获取appkey对应的私钥
func (p *AppkeyService) GetSecretKey(appkey string) (string, bool) {
	item, ok := p.GetAppKeyItem(appkey)
	if !ok {
		return "", false
	}
	return item.SecretKey, true
}
