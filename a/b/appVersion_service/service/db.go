package service

import (
	"github.com/jinzhu/gorm"
	"galaxyotc/common/model"
)

type db struct {
	*gorm.DB
}

func newDB() *db {
	return &db{model.DB}
}

const (
	WaitingApproved = 1
	WaitingPay = 2
)
