package model

import "time"

type AppVersion struct {
	ID          uint64    `gorm:"primary_key" json:"id"`
	Name        string    `gorm:"size:500" json:"name"` // 名称
	AppType     int32     `json:"app_type"`             // app的类型 1苹果 2安卓
	Url         string    `json:"url"`                  // app 所在的路径
	Version     string    `json:"version"`              //app的版本号
	Status      int32     `json:"status"`               // 状态标识：-->0表示没有更新，1表示有更新（可选择更新）；2表示强制更新
	Enable      int32     `json:"enable"`               // 0启用 1禁用
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	CreateUser  uint64    `json:"create_user"`
	UpdatedUser uint64    `json:"updated_user"`
	Remark      string    `json:"remark"` //app更新功能的说明

}
