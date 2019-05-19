package model

import (
	"crypto/md5"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
	"math/rand"
	"strconv"
	"time"
)

// User 用户
type User struct {
	ModelBase
	Name                       string    `json:"name"`                                     // 用户名
	AreaCode                   string    `json:"area_code"`                                // 地区码
	Mobile                     string    `gorm:"index; size:20" json:"-"`                  // 手机
	Email                      string    `gorm:"index; size:50" json:"-"`                  // 邮箱
	Password                   string    `json:"-"`                                        // 密码
	Sex                        int32     `json:"sex"`                                      // 性别
	AvatarURL                  string    `json:"avatar_url"`                               // 头像
	Status                     int32     `json:"status"`                                   // 状态
	InternalAddress            string    `json:"internal_address"`                         // 内部私链地址
	LastLogin                  time.Time `json:"last_login"`                               // 最后登录时间
	ParentId                   uint64    `json:"parent_id"`                                // 上级用户ID
	UserType                   int32     `json:"user_type"`                                // 用户类型
	DiscountRate               float64   `gorm:"type:decimal(19,5)"  json:"discount_rate"` // 折扣率
	IDCardType                 int32     `json:"id_card_type"`                             // 身份照或护照
	IDCardNo                   string    `gorm:"size:50" json:"-"`                         // 身份证号
	IsRealName                 bool      `gorm:"default:false" json:"is_real_name"`        // 是否实名验证
	IsUpdateName               bool      `gorm:"default:false" json:"is_update_name"`      // 是否改过用户名
	RealnameVerificationStatus int32     `json:"realname_verification_status"`             // 实名验证状态
	ReferralCode               string    `gorm:"size:20" json:"referral_code"`             // 用户推荐码
	TradingMethods             string    `json:"trading_methods"`                          // 用户交易方式
	DirectInvitedCount         int32     `json:"direct_invited_count"`                     // 一级用户推广数
	ImToken                    string    `json:"im_token"`                                 // IM服务端令牌
	SnowflakeID                uint64    `gorm:"index;" json:"snowflake_id"`               // 雪花算法ID返给前端用
	EosCode                    string    `json:"eos_code"`                                 // Eos代码
}

// CheckPassword 验证密码是否正确
func (user *User) CheckPassword(password string) bool {
	if password == "" || user.Password == "" {
		return false
	}
	return user.EncryptPassword(password, user.Salt()) == user.Password
}

// Salt 每个用户都有一个不同的盐
func (user *User) Salt() string {
	var userSalt string
	if user.Password == "" {
		userSalt = strconv.Itoa(int(time.Now().Unix()))
	} else {
		userSalt = user.Password[0:10]
	}
	return userSalt
}

// EncryptPassword 给密码加密
func (user *User) EncryptPassword(password, salt string) (hash string) {
	password = fmt.Sprintf("%x", md5.Sum([]byte(password)))
	hash = salt + password + viper.GetString("server.pass_salt")
	hash = salt + fmt.Sprintf("%x", md5.Sum([]byte(hash)))
	return
}

func (user *User) Signin() (string, *BaseUserInfo, error) {
	//生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
	})

	tokenString, err := token.SignedString([]byte(viper.GetString("server.token_secret")))
	if err != nil {
		return "", nil, err
	}

	UpdatesInfo := make(map[string]interface{})

	// 设置当前登录时间为最后一次登录时间
	UpdatesInfo["last_login"] = time.Now().Local()

	// 如果用户邀请码为空时，通过用户ID生成
	if user.ReferralCode == "" {
		UpdatesInfo["referral_code"] = IDToReferralCode(user.ID)
	}

	/*
	if user.ImToken == "" {
		sex := int8(1)
		if user.Sex == UserSexFemale {
			sex = 2
		}

		props := make(map[string]interface{})
		ex := make(map[string]interface{})

		token, err := service.GetImService().UserRegister(user.ID, user.Name, props, user.AvatarURL, user.Email, "", user.Mobile, sex, ex)
		if err == nil {
			UpdatesInfo["im_token"] = token
		}
	}*/

	// 更新用户信息
	if err := DB.Model(&user).Updates(&UpdatesInfo).Error; err != nil {
		return "", nil, err
	}

	// 保存到Redis中
	if err := UserToRedis(*user); err != nil {
		return "", nil, err
	}

	baseInfo := &BaseUserInfo{
		ID:                         user.SnowflakeID,
		Name:                       user.Name,
		AvatarURL:                  user.AvatarURL,
		UserType:                   user.UserType,
		IsRealName:                 user.IsRealName,
		RealnameVerificationStatus: user.RealnameVerificationStatus,
		IsEmail:                    user.Email != "",
		IsMobil:                    user.Mobile != "",
		ReferralCode:               user.ReferralCode,
		ImToken:                    user.ImToken,
	}

	return tokenString, baseInfo, nil
}

// UserFromRedis 从Redis中取出用户信息
func UserFromRedis(userID uint64) (User, error) {
	loginUser := fmt.Sprintf("%s%d", LoginUser, userID)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	userBytes, err := redis.Bytes(RedisConn.Do("GET", loginUser))
	if err != nil {
		fmt.Println(err)
		return User{}, errors.New("未登录")
	}
	var user User
	bytesErr := json.Unmarshal(userBytes, &user)
	if bytesErr != nil {
		fmt.Println(bytesErr)
		return user, errors.New("未登录")
	}
	return user, nil
}

// UserToRedis 将用户信息存到Redis
func UserToRedis(user User) error {
	userBytes, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
		return errors.New("error")
	}
	loginUserKey := fmt.Sprintf("%s%d", LoginUser, user.ID)

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	if _, redisErr := RedisConn.Do("SET", loginUserKey, userBytes, "EX", viper.GetInt("server.token_max_age")); redisErr != nil {
		fmt.Println("redis set failed: ", redisErr.Error())
		return errors.New("error")
	}
	return nil
}

var (
	// 自定义进制
	baseByte = []byte("HVE8S2DZX9C7P5IK3MJUFR4WYLTN6BGQ")
	// 补位字符
	suffixByte = byte('A')
	// 进制长度
	binLength = len(baseByte)
	// 邀请码最小长度
	codeLength = 6
)

// 根据ID生成邀请码
func IDToReferralCode(userID uint64) string {
	buf := make([]byte, binLength)

	charPos := binLength

	id := int(userID)
	// 当id除以数组长度结果大于0，则进行取模操作，并以取模的值作为数组的坐标获得对应的字符
	for id/binLength > 0 {
		index := (id % binLength)
		charPos --
		buf[charPos] = baseByte[index]
		id /= binLength;
	}

	charPos --
	buf[charPos] = baseByte[(id % binLength)]
	// 将字符数组转化为字符串
	result := buf[charPos:]

	// 长度不足指定长度则随机补全
	l := len(result);
	//s :=make([]byte, 0)
	if (l < codeLength) {
		result = append(result, suffixByte)
		// 去除SUFFIX_CHAR本身占位之后需要补齐的位数
		for i := 0; i < codeLength-l-1; i++ {
			result = append(result, baseByte[rand.Intn(binLength)])
		}
	}

	return string(result)
}

// 根据邀请码推回ID
func ReferralCodeToID(code string) uint64 {
	if len(code) != 6 {
		return 0
	}

	byteCode := []byte(code)
	var result uint64
	for i := 0; i < len(byteCode); i++ {
		index := 0;
		for j := 0; j < binLength; j++ {
			if byteCode[i] == baseByte[j] {
				index = j
				break
			}
		}

		if byteCode[i] == suffixByte {
			break;
		}

		if i > 0 {
			result = result*uint64(binLength) + uint64(index)
		} else {
			result = uint64(index)
		}
	}

	return result
}

const (
	// 正常
	UserStatusNormal = 1
	// 已冻结
	UserStatusFrozen = 2
)

const (
	// 正常
	UserStatusNormalString = "正常"
	// 已冻结
	UserStatusFrozenString = "已冻结"
)

func UserStatusString(userType int32) (statuesString string) {
	switch userType {
	case UserStatusNormal:
		statuesString = UserStatusNormalString
	case UserStatusFrozen:
		statuesString = UserStatusFrozenString
	}
	return
}

const (
	// UserSexMale 男
	UserSexMale = 0

	// UserSexFemale 女
	UserSexFemale = 1

	// MaxUserNameLen 用户名的最大长度
	MaxUserNameLen = 20

	// MinUserNameLen 用户名的最小长度
	MinUserNameLen = 4

	// MaxPassLen 密码的最大长度
	MaxPassLen = 20

	// MinPassLen 密码的最小长度
	MinPassLen = 6
)

const (
	//身份证
	IdCardType = 1
	//护照
	Passport = 2
)

const (
	IdCardTypeString = "身份证"
	PassportString   = "护照"
)

func CardTypeString(cardType int32) (cardTypeString string) {
	switch cardType {
	case IdCardType:
		cardTypeString = IdCardTypeString
	case Passport:
		cardTypeString = PassportString
	}
	return
}

const (
	//普通会员
	RegularMembers = 1
	//天使合伙人
	AngelPartner = 2
	//创世合伙人
	FoundingPartner = 3
)

const (
	RegularMembersString  = "普通会员"
	AngelPartnerString    = "天使合伙人"
	FoundingPartnerString = "创世合伙人"
)

const (
	EmailType  = 1
	MobileType = 2
)

const (
	RealnameVerificationStatus_NotVerify = 0 // 未认证
	RealnameVerificationStatus_Verifying = 1 // 认证中
	RealnameVerificationStatus_Approved  = 2 // 认证通过
	RealnameVerificationStatus_Reject    = 3 // 认证不通过
)

func UserTypeString(userType int32) (userTypeString string) {
	switch userType {
	case RegularMembers:
		userTypeString = RegularMembersString
	case AngelPartner:
		userTypeString = AngelPartnerString
	case FoundingPartner:
		userTypeString = FoundingPartnerString
	}
	return
}

type BaseUserInfo struct {
	ID                         uint64 `json:"id"`           		// 用户ID
	AreaCode                   string `json:"area_code"`    		// 国家地区编号
	Name                       string `json:"name"`         		// 用户名
	AvatarURL                  string `json:"avatar_url"`   		// 用户头像
	UserType                   int32  `json:"user_type"`    		// 用户类型
	IsRealName                 bool   `json:"is_real_name"` 		// 是否实名
	RealnameVerificationStatus int32  `json:"realname_verification_status"`
	IsEmail                    bool   `json:"is_email"`      		// 是否绑定邮箱
	IsMobil                    bool   `json:"is_mobil"`      		// 是否绑定手机
	IsUpdateName               bool   `json:"is_update_name"`      	// 是否改过用户名
	ReferralCode               string `json:"referral_code"` 		// 邀请码
	ImToken                    string `json:"im_token"`      		// IM服务端令牌
}

//用户信息
type UserInfo struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	UserType       int32     `json:"user_type"`
	UserTypeString string    `json:"user_type_string"`
	AvatarURL      string    `json:"avatar_url"`
	CreatedAt      time.Time `json:"create_time"`
}

type InviteeInfo struct {
	ID                 uint64 `json:"id"`
	Name               string `json:"name"`
	UserType           int32  `json:"user_type"`
	AvatarUrl          string `json:"avatar_url"`
	DirectInvitedCount int32  `json:"direct_invited_count"`
}

//用户后台信息
type UserBaseInfo struct {
	ID                 uint64    `json:"id"`                   // 用户ID
	AreaCode           string    `json:"area_code"`            // 区域
	SnowflakeID        uint64    `json:"snowflake_id"`         // 用户唯一随机码
	Name               string    `json:"name"`                 // 用户名
	Mobile             string    `json:"mobile"`               // 手机
	Email              string    `json:"email"`                // 邮箱
	AvatarURL          string    `json:"avatar_url"`           // 头像
	Status             int32     `json:"status"`               // 状态
	StatusString       string    `json:"status_string"`        // 状态
	InternalAddress    string    `json:"internal_address"`     // 内部私链地址
	LastLogin          time.Time `json:"last_login"`           // 最后登录时间
	ParentName         string    `json:"parent_name"`          // 上级用户ID
	UserType           string    `json:"user_type"`            // 用户类型
	DiscountRate       float64   `json:"discount_rate"`        // 折扣率
	IDCardType         int32     `json:"id_card_type"`         // 身份照或护照
	IDCardNo           string    `json:"id_card_no"`           // 身份证号
	IsRealName         bool      `json:"is_real_name"`         // 是否实名验证
	ReferralCode       string    `json:"referral_code"`        // 用户推荐码
	DirectInvitedCount int32     `json:"direct_invited_count"` //一级用户推广数
	CreatedAt          time.Time `json:"created_at"`           //创建时间
}

//用户后台信息
type UserExportInfo struct {
	ID                 uint64  `json:"id"`                   // 用户ID
	AreaCode           string  `json:"area_code"`            // 区域
	Code               string  `json:"code"`                 // 用户唯一随机码
	Name               string  `json:"name"`                 // 用户名
	Mobile             string  `json:"mobile"`               // 手机
	Email              string  `json:"email"`                // 邮箱
	AvatarURL          string  `json:"avatar_url"`           // 头像
	InternalAddress    string  `json:"internal_address"`     // 内部私链地址
	LastLogin          string  `json:"last_login"`           // 最后登录时间
	ParentName         string  `json:"parent_name"`          // 上级用户ID
	UserType           string  `json:"user_type"`            // 用户类型
	DiscountRate       float64 `json:"discount_rate"`        // 折扣率
	IDCardType         string  `json:"id_card_type"`         // 身份照或护照
	IDCardNo           string  `json:"id_card_no"`           // 身份证号
	IsRealName         string  `json:"is_real_name"`         // 是否实名验证
	ReferralCode       string  `json:"referral_code"`        // 用户推荐码
	DirectInvitedCount int32   `json:"direct_invited_count"` //一级用户推广数
	CreatedAt          string  `json:"created_at"`           //创建时间
	StatusString       string  `json:"status_string"`        // 状态字符串
}
