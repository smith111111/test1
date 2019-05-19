package model

import "time"

// 公告
type Notice struct {
	ID           uint       `gorm:"primary_key" json:"id"`
	Name         string     `json:"name"`
	Summary      string     `json:"summary"`
	Content      string     `json:"content"`
	Status       int        `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at"`
	ApprovedUser uint       `json:"approved_user"`
}

// 公告信息
type NoticeListInfo struct {
	ID        uint      `json:"id"`
	Status    int       `json:"status"`
	Name      string    `json:"name"`
	Summary   string    `json:"summary"`
	Content      string     `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	CreatedAtString string `json:"created_at_string"`
	ApprovedAt   *time.Time `json:"approved_at"`
}

type NoticeInfo struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	// 审核中
	Notice_Approving = 0

	// 审核通过
	Notice_ApproveSuccess = 1

	// 审核未通过
	Notice_ApproveFail = 2
)
