package service

import (


	pb "galaxyotc/common/proto/backend/user"
	"github.com/jinzhu/gorm"


	"galaxyotc/common/model"
	"galaxyotc/common/utils"
	"galaxyotc/common/log"
)

type db struct {
	*gorm.DB
}

func newDB() *db {
	return &db{model.DB}
}

// 获取用户信息
func (db *db) GetUser(id uint64) (*pb.UserInfo, error) {
	var user pb.UserInfo
	if err := db.Table("users").
		Select("id, name, area_code, mobile, email, avatar_url, status, internal_address, parent_id, user_type, discount_rate, is_real_name, referral_code, trading_methods").
		Where("id = ?", id).Scan(&user).Error; err != nil {
		log.Errorf("api-GetUser-error: %s", err.Error())
		return nil, err
	}

	if user.Email != "" {
		user.Email = utils.TextReplaceForEmail(user.Email)
	}

	if user.Mobile != "" {
		user.Mobile = utils.TextReplace(user.Mobile[len(user.AreaCode):])
	}

	return &user, nil
}

// 根据内部地址获取用户信息
func (db *db) GetUserByInternalAddress(internalAddress string) (*pb.UserInfo, error) {
	var user pb.UserInfo
	if err := db.Table("users").
		Select("id, name, area_code, mobile, email, avatar_url, status, internal_address, parent_id, user_type, discount_rate, is_real_name, referral_code, trading_methods").
		Where("internal_address = ?", internalAddress).Scan(&user).Error; err != nil {
		log.Errorf("api-GetUser-error: %s", err.Error())
		return nil, err
	}

	if user.Email != "" {
		user.Email = utils.TextReplaceForEmail(user.Email)
	}

	if user.Mobile != "" {
		user.Mobile = utils.TextReplace(user.Mobile[len(user.AreaCode):])
	}

	return &user, nil
}

// 检查用户是否是否存在
func (db *db) IsExist(input string) (bool, error) {
	// 判断需要绑定的该手机是否已经注册
	var user model.User
	err := db.Where("email = ?", input).Or("mobile = ?", input).First(&user).Error
	if err == nil {
		return true, nil
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return false, nil
	} else {
		log.Errorf("api-IsExist-error: %s", err.Error())
		return false, err
	}
}