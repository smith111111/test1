package model

import "time"

type OperateMenu struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	Name      string    `gorm:"size:100" json:"name"` // 名称
	Status    int32     `json:"status"`               // 状态
	Remark    string    `json:"remark"`               // 备注
	Sort      int       `json:"sort"`                 // 排序
	CreatedAt time.Time `json:"created_at"`           // 创建时间
	UpdatedAt time.Time `json:"updated_at"`           // 更新时间
}


const (
	OperateMenuStatusEnableInt = 0
	OperateMenuStatusDisableInt= 1
)

const (
	OperateMenuStatusEnableString = "启用"
	OperateMenuStatusDisableString= "禁用"
)
