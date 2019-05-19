package user

import (
	"fmt"
	"time"
	"math"
	"strings"
	"net/http"

	"galaxyotc/common/net"
	//"galaxyotc/common/data"
	"galaxyotc/common/utils"
	"galaxyotc/common/log"
	"galaxyotc/common/model"
	//searchService "galaxyotc/common/service/search_service"

	wi "galaxyotc/wallet-interface"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"

	"galaxyotc/gc_services/user_service/api"
	//"galaxyotc/gc_services/user_service/client"
	"github.com/gao88/invite_code"
	"strconv"
)

// 登录返回信息
func UserSignin(user *model.User) (string, *model.BaseUserInfo, error) {
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
		UpdatesInfo["referral_code"] = model.IDToReferralCode(user.ID)
	}

	/*if user.ImToken == "" {
		sex := int8(1)
		if user.Sex == model.UserSexFemale {
			sex = 2
		}

		props := make(map[string]interface{})
		ex := make(map[string]interface{})

		token, err := api.ImApi.UserRegister(user.ID, user.Name, props, user.AvatarURL, user.Email, "", user.Mobile, sex, ex)
		if err == nil {
			UpdatesInfo["im_token"] = token
		}
	}*/

	// 更新用户信息
	if err := model.DB.Model(&user).Updates(&UpdatesInfo).Error; err != nil {
		return "", nil, err
	}

	// 保存到Redis中
	if err := model.UserToRedis(*user); err != nil {
		return "", nil, err
	}

	baseInfo := &model.BaseUserInfo{
		ID:                       	user.SnowflakeID,
		Name:                       user.Name,
		AvatarURL:                  user.AvatarURL,
		UserType:                   user.UserType,
		IsRealName:                 user.IsRealName,
		RealnameVerificationStatus: user.RealnameVerificationStatus,
		IsEmail:                    user.Email != "",
		IsMobil:                    user.Mobile != "",
		ReferralCode:               user.ReferralCode,
		ImToken:                    user.ImToken,
		IsUpdateName: 				user.IsUpdateName,
	}

	return tokenString, baseInfo, nil
}

type signinReq struct {
	Input    string `json:"input" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// 用户登录
func Signin(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req signinReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-Signin-Error: %s", err.Error())
		SendErrJSON("参数有误", c)
		return
	}

	var user model.User
	if err := model.DB.Where("email = ?", req.Input).Or("mobile = ?", req.Input).First(&user).Error; err != nil {
		log.Errorf("User-Signin-Error: %s", err.Error())
		SendErrJSON("用户不存在", c)
		return
	}

	if result := user.CheckPassword(req.Password); !result {
		SendErrJSON("账号或密码错误", c)
		return
	}

	token, baseInfo, err := UserSignin(&user)
	if err != nil {
		log.Errorf("User-Signin-Error: %s", err.Error())
		SendErrJSON("用户登录失败", c)
		return
	}

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	msgMap := make(map[string]interface{})
	//	msgMap["t1"] = "v1"
	///	api.PushApi.SendMsg("notification", fmt.Sprintf("%d", user.ID), data.APPID_OTC, "登录成功", "登录成功测试通过!", msgMap, 1)
	//}()

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"token":     token,
			"base_info": baseInfo,
		},
	})
}

type signupReq struct {
	Input        	string 	`json:"input" binding:"required"`
	Password     	string 	`json:"password" binding:"required"`
	Captcha      	string 	`json:"captcha" binding:"required"`
	AreaCode     	string 	`json:"area_code"`
	ReferralCode 	string 	`json:"referral_code"`
	Type 			int 	`json:"type" binding:"required"`
}

// 用户注册
func Signup(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req signupReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-Signup-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if len(req.Password) < model.MinPassLen || len(req.Password) > model.MaxPassLen {
		SendErrJSON(fmt.Sprintf("密码长度必须大于%d小于%d", model.MinPassLen, model.MaxPassLen), c)
		return
	}

	var (
		newUser model.User
		CaptchaKey string
	)

	switch req.Type {
	case model.EmailType:
		// 去除空格
		email := strings.TrimSpace(req.Input)
		// 校验是否是有效的邮箱格式
		if !utils.RegexpEmail(email) {
			SendErrJSON("无效的邮箱格式", c)
			return
		}

		CaptchaKey = fmt.Sprintf("%s%s", model.EmailCaptcha, email)

		if !model.DB.Model(model.User{}).Where("email = ?", req.Input).Find(&model.User{}).RecordNotFound() {
			SendErrJSON(fmt.Sprintf("邮箱%s已经注册", req.Input), c)
			return
		}

		newUser.Email = req.Input
	case model.MobileType:
		// 如果是手机注册需要获取地区码
		if req.AreaCode == "" {
			SendErrJSON("地区码不能为空", c)
			return
		}

		// 去除空格
		mobile := strings.TrimSpace(req.Input)
		// 校验是否是有效的手机格式
		if !utils.RegexpMobile(mobile[len(req.AreaCode):]) {
			SendErrJSON("无效的手机格式", c)
			return
		}

		CaptchaKey = fmt.Sprintf("%s%s", model.MobileCaptcha, mobile)

		if !model.DB.Model(model.User{}).Where("mobile = ?", req.Input).Find(&model.User{}).RecordNotFound() {
			SendErrJSON(fmt.Sprintf("手机%s已经注册", req.Input), c)
			return
		}

		newUser.Mobile = req.Input
		newUser.AreaCode = req.AreaCode
	default:
		SendErrJSON("无效的注册类型", c)
	}

	if result, err := model.VerifyCaptcha(req.Captcha, CaptchaKey); !result {
		log.Errorf("User-Signup-Error: %s", err.Error())
		SendErrJSON("验证码错误或过期", c)
		return
	}

	// 如果邀请码不为空，根据邀请码逆推得出注册用户的上级
	var parentID uint64
	if strings.TrimSpace(req.ReferralCode) != "" {
		parentID = model.ReferralCodeToID(req.ReferralCode)

		var parentUser model.User
		if err := model.DB.First(&parentUser, parentID).Error; err != nil {
			log.Errorf("User-Signup-Error: %s", err.Error())
			parentID = 0
		}
	}

	// 使用雪花算法生成用户ID
	newUser.SnowflakeID = utils.GetUidMgrService().GetNextUserId()
	newUser.EosCode = invite_code.GetInviteCode(newUser.SnowflakeID)
	sfStr := strconv.FormatUint(newUser.SnowflakeID, 10)
	// 用户名为雪花算法的最后8位
	newUser.Name = "gc_" + sfStr[len(sfStr)-8:]
	newUser.Password = newUser.EncryptPassword(req.Password, newUser.Salt())
	newUser.UserType = model.RegularMembers
	newUser.Status = model.UserStatusNormal
	newUser.Sex = model.UserSexMale
	newUser.AvatarURL = ""
	// 给新创建的用户分配一个内部私链地址
	address, err := api.PrivateWalletApi.NewAddress(int32(wi.EXTERNAL))
	if err != nil {
		log.Errorf("User-Signup-Error: %s", err.Error())
		SendErrJSON("创建新用户失败", c)
		return
	}
	newUser.InternalAddress = address
	newUser.LastLogin = time.Now().Local()
	newUser.ParentId = parentID

	if err := model.DB.Create(&newUser).Error; err != nil {
		log.Errorf("User-Signup-Error: %s", err.Error())
		SendErrJSON("创建新用户失败", c)
		return
	}

	// 给新创建用户的内部地址分配一些ETH
	valueWei := utils.ToWei(int64(100), 18).String()
	if _, err := api.PrivateWalletApi.Transfer(newUser.InternalAddress, valueWei); err != nil {
		log.Errorf("User-Signup-Error: %s", err.Error())
	}

	//添加进搜索
	//{
	//	userInfo := &searchService.SearchUserInfo{
	//		Id: newUser.ID,
	//		Name: newUser.Name,
	//		Mobile: newUser.Mobile,
	//		Email: newUser.Email,
	//	}
		//client.SearchServiceClient.AsyncAddUserInfo(newUser.ID, userInfo)
	//}

	//统计推荐人一共推荐了多少个一级注册用户
	if parentID > 0 {
		go func() {
			var (
				DirectInvitedCount uint
			)
			var parentUser model.User
			tx := model.DB.Begin()
			if err := tx.First(&parentUser, parentID).Error; err != nil {
				log.Errorf("User-Signup-Error: %s", err.Error())
				tx.Rollback()
				return
			}

			if err := tx.Table("users").Where("parent_id = ?", parentID).Count(&DirectInvitedCount).Error; err != nil {
				log.Errorf("User-Signup-Error: %s", err.Error())
				tx.Rollback()
				return
			}

			if err := tx.Model(&parentUser).Update("direct_invited_count", DirectInvitedCount).Error; err != nil {
				log.Errorf("User-Signup-Error: %s", err.Error())
				tx.Rollback()
				return
			}

			tx.Commit()
		}()
	}

	// 注册成功后就自动登录
	token, baseInfo, err := UserSignin(&newUser)
	if err != nil {
		log.Errorf("User-Signup-Error: %s", err.Error())
		SendErrJSON("自动登录失败, 请重新尝试", c)
		return
	}

	//go func() {
	//	msgMap := make(map[string]interface{})
	//	msgMap["t1"] = "v1"
	//	api.PushApi.SendMsg("notification", fmt.Sprintf("%d", newUser.ID), data.APPID_OTC, "注册成功", "注册成功测试通过!", msgMap, 1)
	//}()

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"token":     token,
			"base_info": baseInfo,
		},
	})
}

// 退出登录
func Signout(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	userInter, exists := c.Get("user")
	if exists {
		user := userInter.(model.User)

		RedisConn := model.RedisPool.Get()
		defer RedisConn.Close()

		if _, err := RedisConn.Do("DEL", fmt.Sprintf("%s%d", model.LoginUser, user.ID)); err != nil {
			log.Errorf("User-Signout-Error: %s", err.Error())
			SendErrJSON("用户登出失败", c)
			return
		}

		/*if err := api.PushApi.Logout(user.ID, data.APPID_OTC); err != nil {
			log.Errorf("User-Signout-Error: %s", err.Error())
		}*/
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 用户个人信息
func Info(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取当前用户
	userInter, _ := c.Get("user")
	currentUser := userInter.(model.User)

	// 获取用户最新信息
	var user model.User
	if err := model.DB.First(&user, currentUser.ID).Error; err != nil {
		log.Errorf("User-Info-Error: %s", err.Error())
		SendErrJSON("获取用户个人信息失败", c)
		return
	}

	//生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
	})

	tokenString, err := token.SignedString([]byte(viper.GetString("server.token_secret")))
	if err != nil {
		log.Errorf("User-Info-Error: %s", err.Error())
		SendErrJSON("获取用户Token失败", c)
		return
	}

	baseInfo := &model.BaseUserInfo{
		ID:                       	user.SnowflakeID,
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

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"token": tokenString,
			"base_info": baseInfo,
		},
	})
}

type bindMobileOrEmailReq struct {
	Input    string `json:"input" binding:"required"`
	AreaCode string `json:"area_code"`
	Captcha  string `json:"captcha" binding:"required"`
	Type 	 int 	`json:"type" binding:"required"`
}

// 绑定手机或邮箱
func BindMobileOrEmail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req bindMobileOrEmailReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-BindMobileOrEmail-Error: %s", err.Error())
		SendErrJSON("请求参数有误", c)
		return
	}

	// 获取当前用户
	userInter, _ := c.Get("user")
	currentUser := userInter.(model.User)

	// 手机、邮箱、密码属于隐私数据，所以只能通过用户ID进行查询获取用户信息
	var user model.User
	if err := model.DB.First(&user, currentUser.ID).Error; err != nil {
		log.Errorf("User-BindMobileOrEmail-Error: %s", err.Error())
		SendErrJSON(fmt.Sprintf("无效的用户", req.Input), c)
		return
	}

	var CaptchaKey string

	UpdatesInfo := make(map[string]interface{})

	switch req.Type {
	case model.EmailType:
		if user.Email != "" {
			SendErrJSON(fmt.Sprintf("您已绑定了邮箱账号,无需重新绑定", req.Input), c)
			return
		}

		// 去除空格
		email := strings.TrimSpace(req.Input)
		// 校验是否是有效的邮箱格式
		if !utils.RegexpEmail(email) {
			SendErrJSON("无效的邮箱格式", c)
			return
		}

		CaptchaKey = fmt.Sprintf("%s%s", model.EmailCaptcha, email)

		if !model.DB.Model(model.User{}).Where("email = ?", req.Input).Find(&model.User{}).RecordNotFound() {
			SendErrJSON(fmt.Sprintf("邮箱%s已经注册", req.Input), c)
			return
		}

		UpdatesInfo["email"] = req.Input
	case model.MobileType:
		if user.Mobile != "" {
			SendErrJSON(fmt.Sprintf("您已绑定了手机号码,无需重新绑定", req.Input), c)
			return
		}

		// 如果是绑定手机需要获取地区码
		if req.AreaCode == "" {
			SendErrJSON(fmt.Sprintf("地区码不能为空", req.Input), c)
			return
		}

		// 去除空格
		mobile := strings.TrimSpace(req.Input)
		// 校验是否是有效的手机格式
		if !utils.RegexpMobile(mobile[len(req.AreaCode):]) {
			SendErrJSON("无效的手机格式", c)
			return
		}

		CaptchaKey = fmt.Sprintf("%s%s", model.MobileCaptcha, mobile)

		if !model.DB.Model(model.User{}).Where("mobile = ?", req.Input).Find(&model.User{}).RecordNotFound() {
			SendErrJSON(fmt.Sprintf("手机%s已经注册", req.Input), c)
			return
		}

		UpdatesInfo["area_code"] = req.AreaCode
		UpdatesInfo["mobile"] = req.Input
	default:
		SendErrJSON("无效的注册类型", c)
		return
	}

	// 校验验证码
	if result, _ := model.VerifyCaptcha(req.Captcha, CaptchaKey); !result {
		SendErrJSON("验证码错误或过期", c)
		return
	}

	// 更改绑定信息
	if err := model.DB.Model(&user).Updates(UpdatesInfo).Error; err != nil {
		log.Errorf("User-BindMobileOrEmail-Error: %s", err.Error())
		SendErrJSON("绑定失败", c)
		return
	}

	// 更新搜索
	//client.SearchServiceClient.AsyncUpdateUserInfo(user.ID, UpdatesInfo)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

type checkPasswordReq struct {
	Password string `json:"password" binding:"required"`
}

func CheckPassword(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req checkPasswordReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-ResetPassword-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if len(req.Password) < model.MinPassLen || len(req.Password) > model.MaxPassLen {
		SendErrJSON(fmt.Sprintf("密码长度必须大于%d小于%d", model.MinPassLen, model.MaxPassLen), c)
		return
	}

	// 获取当前用户
	userInter, _ := c.Get("user")
	currentUser := userInter.(model.User)

	// 手机、邮箱、密码属于隐私数据，所以只能通过用户ID进行查询获取用户信息
	var user model.User
	if err := model.DB.First(&user, currentUser.ID).Error; err != nil {
		log.Errorf("User-ResetPassword-Error: %s", err.Error())
		SendErrJSON("无效的用户", c)
		return
	}

	if !user.CheckPassword(req.Password) {
		SendErrJSON("密码不正确", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

type resetPasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// 重置密码
func ResetPassword(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req resetPasswordReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-ResetPassword-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if len(req.NewPassword) < model.MinPassLen || len(req.NewPassword) > model.MaxPassLen {
		SendErrJSON(fmt.Sprintf("密码长度必须大于%d小于%d", model.MinPassLen, model.MaxPassLen), c)
		return
	}

	// 获取当前用户
	userInter, _ := c.Get("user")
	currentUser := userInter.(model.User)

	// 手机、邮箱、密码属于隐私数据，所以只能通过用户ID进行查询获取用户信息
	var user model.User
	if err := model.DB.First(&user, currentUser.ID).Error; err != nil {
		log.Errorf("User-ResetPassword-Error: %s", err.Error())
		SendErrJSON("无效的用户", c)
		return
	}

	if !user.CheckPassword(req.OldPassword) {
		SendErrJSON("原密码不正确", c)
		return
	}

	if err := model.DB.Model(&user).Update("password", user.EncryptPassword(req.NewPassword, user.Salt())).Error; err != nil {
		log.Errorf("User-ResetPassword-Error: %s", err.Error())
		SendErrJSON("重置密码失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

type resetPasswordForGetBackReq struct {
	Input       string 	`json:"input" binding:"required"`
	NewPassword string 	`json:"new_password" binding:"required"`
	Captcha     string 	`json:"captcha" binding:"required"`
	Type 		int 	`json:"type" binding:"required"`
}

// 忘记密码
func ResetPasswordForGetBack(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req resetPasswordForGetBackReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-ResetPasswordForGetBack-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	if len(req.NewPassword) < model.MinPassLen || len(req.NewPassword) > model.MaxPassLen {
		SendErrJSON(fmt.Sprintf("密码长度必须大于%d小于%d", model.MinPassLen, model.MaxPassLen), c)
		return
	}

	var (
		sql         string
		CaptchaKey  string
	)

	switch req.Type {
	case model.EmailType:
		sql = "email = ?"
		CaptchaKey = fmt.Sprintf("%s%s", model.EmailCaptcha, req.Input)
	case model.MobileType:
		sql = "mobile = ?"
		CaptchaKey = fmt.Sprintf("%s%s", model.MobileCaptcha, req.Input)
	default:
		SendErrJSON("无效的参数", c)
		return
	}

	var user model.User
	if err := model.DB.Where(sql, req.Input).First(&user).Error; err != nil {
		log.Errorf("User-ResetPasswordForGetBack-Error: %s", err.Error())
		SendErrJSON("无效的用户", c)
		return
	}

	// 校验验证码
	if result, _ := model.VerifyCaptcha(req.Captcha, CaptchaKey); !result {
		SendErrJSON("验证码错误或过期", c)
		return
	}

	if err := model.DB.Model(&user).Update("password", user.EncryptPassword(req.NewPassword, user.Salt())).Error; err != nil {
		log.Errorf("User-ResetPasswordForGetBack-Error: %s", err.Error())
		SendErrJSON("找回密码失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 修改用户头像
func UpdateAvatar(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	uploadData, err := net.Upload(c, net.UserAvatar)
	if err != nil {
		log.Errorf("User-UpdateAvatar-Error: %s", err.Error())
		SendErrJSON("上传头像失败", c)
		return
	}
	avatarUrl := uploadData["url"].(string)

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	if err := model.DB.First(&user, user.ID).Error; err != nil {
		log.Errorf("User-UpdateAvatar-Error: %s", err.Error())
		SendErrJSON("无效的用户", c)
		return
	}

	if err := model.DB.Model(&user).Update("avatar_url", avatarUrl).Error; err != nil {
		log.Errorf("User-UpdateAvatar-Error: %s", err.Error())
		SendErrJSON("更新用户头像失败", c)
	}

	if err := model.UserToRedis(user); err != nil {
		log.Errorf("User-UpdateAvatar-Error: %s", err.Error())
		SendErrJSON("更新用户头像失败", c)
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

type updateNameReq struct {
	Name	string 	`json:"name"`
}

// 修改用户姓名
func UpdateName(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req updateNameReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.Errorf("User-UpdateName-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	// 获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	if err := model.DB.First(&user, user.ID).Error; err != nil {
		log.Errorf("User-UpdateName-Error: %s", err.Error())
		SendErrJSON("无效的用户", c)
		return
	}

	// 一个用户只能改一次用户名
	if user.IsUpdateName {
		SendErrJSON("已超过改名次数", c)
		return
	}

	// 用户名唯一
	if !model.DB.Where(model.User{Name: req.Name}).First(model.User{}).RecordNotFound() {
		SendErrJSON("该用户名已被使用", c)
		return
	}

	if err := model.DB.Model(&user).Update(model.User{Name: req.Name, IsUpdateName: true}).Error; err != nil {
		log.Errorf("User-UpdateName-Error: %s", err.Error())
		SendErrJSON("更新用户昵称失败", c)
		return
	}

	//更新搜索
	//{
	//	userInfo := make(map[string]interface{})
	//	userInfo["name"] = req.Name
	//	client.SearchServiceClient.AsyncUpdateUserInfo(user.ID, userInfo)
	//}

	if err := model.UserToRedis(user); err != nil {
		log.Errorf("User-UpdateName-Error: %s", err.Error())
		SendErrJSON("更新用户昵称失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 获取我的邀请会员
func UserInvitees(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	offset := (page - 1) * size

	var users []*model.User

	baseQuery := model.DB.Model(&model.User{}).Where("parent_id = ?", user.ID).Order("created_at DESC")

	var totalCount, regularMembersCount, angelPartnerCount int64

	// 获取总邀请数
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("User-UserInvitees-Error: %s", err.Error())
		SendErrJSON("获取我的邀请列表失败", c)
		return
	}

	invitees := []*model.InviteeInfo{}

	if totalCount > 0 {
		//普通会员数
		if err := baseQuery.Where("user_type = ?", model.RegularMembers).Count(&regularMembersCount).Error; err != nil {
			log.Errorf("User-UserInvitees-Error: %s", err.Error())
			SendErrJSON("获取我的邀请会员列表失败", c)
			return
		}
		//天使合伙人数
		if err := baseQuery.Where("user_type = ?", model.AngelPartner).Count(&angelPartnerCount).Error; err != nil {
			log.Errorf("User-UserInvitees-Error: %s", err.Error())
			SendErrJSON("获取我的邀请会员列表失败", c)
			return
		}

		if err := baseQuery.Offset(offset).Limit(size).Find(&users).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("UserInvitees Error: %s", err.Error())
			SendErrJSON("获取我的邀请会员列表失败", c)
			return
		}

		for _, user := range users {
			name := user.Name
			//if user.Mobile != "" {
			//	name = utils.TextReplace(user.Mobile[len(user.AreaCode):])
			//} else if user.Email != "" {
			//	name = utils.TextReplaceForEmail(user.Email)
			//}

			invitees = append(invitees, &model.InviteeInfo{
				ID:               	user.ID,
				Name:               name,
				UserType:           user.UserType,
				AvatarUrl:          user.AvatarURL,
				DirectInvitedCount: user.DirectInvitedCount,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"invitees":            	invitees,
			"regularMembersCount": 	regularMembersCount,
			"angelPartnerCount":   	angelPartnerCount,
			"pageNo":     			page,
			"pageSize":   			size,
			"totalPage": 	 		math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": 			totalCount,
		},
	})
}

// 获取我的收益
func UserCommissions(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	page, size := net.GetPageAndSize(c)
	offset := (page - 1) * size

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var (
		commissions     []*model.CommissionDistribution
		totalCount 		int64
	)

	commissionList := []*model.CommissionInfo{}

	baseQuery := model.DB.Model(&model.CommissionDistribution{}).Where("user_id = ?", user.ID).Order("created_at")

	// 获取总数
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("User-UserCommissions-Error: %s", err.Error())
		SendErrJSON("获取我的收益失败", c)
		return
	}

	if totalCount > 0 {
		// 获取列表
		if err := baseQuery.Offset(offset).Limit(size).Find(&commissions).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("User-UserCommissions-Error: %s", err.Error())
			SendErrJSON("获取我的收益失败", c)
			return
		}

		for _, commission := range commissions {
			commissionList = append(commissionList, &model.CommissionInfo {
				BusinessType:       commission.BusinessType,
				Amount:       		commission.Amount,
				CreatedAt:       	commission.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H {
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H {
			"commissions":      commissionList,
			"pageNo":     		page,
			"pageSize":   		size,
			"totalPage": 	 	math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": 		totalCount,
		},
	})
}