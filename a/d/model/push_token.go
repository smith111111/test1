package model

// 用户对应push_token表
type PushToken struct {
	UserId      uint64   	`json:"user_id"`             									// 用户ID
	AppId       int32   	`gorm:"type:tinyint(1)" json:"app_id"`               			// 应用标识,1:OTC
	DeviceToken string 		`gorm:"type:varchar(64)" json:"device_token"` 					// 推送token
	Platform    int32   		`gorm:"tinyint(1)" json:"platform"`         					// 所属平台,1:android, 2:ios
	LoginStatus int32   		`gorm:"type:tinyint(1)" json:"login_status"` 					// 登录状态，-1：退出。1登录
	PushType    int32   	`gorm:"type:tinyint(1)" json:"push_type"`       				// 推送系统类型, 0:友盟推送, 1:极光推送, 2:小米推送
	UpdateTime  int64  		`gorm:"column:update_time;type:int(11)" json:"update_time"`   	// 更新时间
}
