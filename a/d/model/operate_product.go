package model

import "time"

type OperateProduct struct {
	ID         uint64    `gorm:"primary_key" json:"id"`
	Name       string    `gorm:"size:100" json:"name"` // 名称
	MenuID     int       `json:"menu_id"`              // 菜单id
	CoverImg   string    `json:"cover_img"`            // 封面图片
	Url        string    `json:"url"`                  // 链接
	Status     int32     `json:"status"`               // 状态
	CategoryID int       `json:"category_id"`          // 当前菜单下页面的分类ID
	Remark     string    `json:"remark"`               // 备注
	Sort       int       `json:"sort"`                 // 排序
	CreatedAt  time.Time `json:"created_at"`           // 创建时间
	UpdatedAt  time.Time `json:"updated_at"`           // 更新时间
	OperateType int `json:"operate_type"`      //操作类型
}


const (
	OperateProductStatusEnableInt = 0
	OperateProductStatusDisableInt= 1
)

const (
	OperateProductStatusEnableString = "启用"
	OperateProductStatusDisableString= "禁用"
)
