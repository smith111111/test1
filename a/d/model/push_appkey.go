package model

// 应用的秘钥管理表
type PushAppKey struct {
	AppId       int8   `gorm:"column:appid;type:tinyint(1)" json:"appid"`                // 应用标识, 1:OTC
	Platform    int8   `gorm:"column:platform;type:tinyint(1)" json:"platform"`          // 所属平台, 1:android, 2:ios
	PushType    int8   `gorm:"column:push_type;type:tinyint(1)" json:"push_type"`        // 推送系统类型, 0:友盟推送, 1:极光推送, 2:小米推送
	AppKey      string `gorm:"column:appkey;type:varchar(64)" json:"appkey"`             // AppKey
	Secret      string `gorm:"column:secret;type:varchar(64)" json:"secret"`             // 应用标识对的秘钥
	PackageName string `gorm:"column:package_name;type:varchar(64)" json:"package_name"` // 客户端包名
}
