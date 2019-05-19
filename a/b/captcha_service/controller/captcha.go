package controller

import (
	"fmt"
	"errors"
	"regexp"
	"net/http"

	"galaxyotc/gc_services/captcha_service/api"
	"galaxyotc/gc_services/captcha_service/client"

	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"
	"galaxyotc/common/net"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

const (
	EmailType = 1
	MobileType = 2
)

// 根据语言选择对应的内容并拼接请求路径然后发送短信
func sendMobileCaptcha(captcha, toMobile, language string) {
	var content string
	// TODO
	switch language {
	case "CN":
		content = fmt.Sprintf(viper.GetString("captcha_content.cn"), captcha)
	case "EN":
		content = fmt.Sprintf(viper.GetString("captcha_content.en"), captcha)
	case "JA":
		content = fmt.Sprintf(viper.GetString("captcha_content.ja"), captcha)
	default:
		content = fmt.Sprintf(viper.GetString("captcha_content.cn"), captcha)
	}

	req, err := http.NewRequest("GET", viper.GetString("hxd.api"), nil)
	if err != nil {
		log.Errorf("sendMobileCaptcha Error: %s", err.Error())
	}

	query := req.URL.Query()
	query.Add("username", viper.GetString("hxd.user"))
	query.Add("password",viper.GetString("hxd.password"))
	query.Add("to", toMobile)
	query.Add("text", content)
	req.URL.RawQuery = query.Encode()

	client.SendSMSMsg(req.URL.String())
}

// 根据语言选择对应的内容然后发送邮件
func sendMailCaptcha(title, captcha, toEmail, language string) {
	var content string
	// TODO
	switch language {
	case "CN":
		content = fmt.Sprintf(viper.GetString("captcha_content.cn"), captcha)
	case "EN":
		content = fmt.Sprintf(viper.GetString("captcha_content.en"), captcha)
	case "JA":
		content = fmt.Sprintf(viper.GetString("captcha_content.ja"), captcha)
	default:
		content = fmt.Sprintf(viper.GetString("captcha_content.cn"), captcha)
	}

	client.SendMail(toEmail, title, content)
}

// 根据类型发送对应的验证码
func sendCaptcha(captchaType int, to string, language string) error {
	// 获取Redis连接
	RedisConn := model.RedisPool.Get()
	defer RedisConn.Close()

	limitKey := model.CaptchaMinuteLimit + to
	limitCount, err := redis.Int64(RedisConn.Do("GET", limitKey))
	if err == nil && limitCount >= model.CaptchaMinuteLimitCount {
		return errors.New("发送验证码操作过于频繁，请先休息一会儿。")
	}

	minuteRemainingTime, _ := redis.Int64(RedisConn.Do("TTL", limitKey))
	if minuteRemainingTime < 0 || minuteRemainingTime > 60 {
		minuteRemainingTime = 60
	}

	if _, err := RedisConn.Do("SET", limitKey, limitCount + 1, "EX", minuteRemainingTime); err != nil {
		log.Errorf("sendCaptcha Error: %s", err.Error())
		return errors.New("服务器出错啦！")
	}

	// 获取新的六位数字验证码
	captcha := utils.NewCaptcha(6, utils.NumberCaptcha)
	captchaMaxAge := viper.GetInt("server.captcha_max_age")

	switch captchaType {
	case EmailType:
		// TODO: Key添加个前缀，用于区别每个场景
		// 拼接Key
		emailCaptcha := fmt.Sprintf("%s%s", model.EmailCaptcha, to)

		// 将验证码保存到Redis数据库中
		if _, err := RedisConn.Do("SET", emailCaptcha, captcha, "EX", captchaMaxAge); err != nil {
			return errors.New("服务器出错啦！")
		}

		go func() {
			sendMailCaptcha("[Galaxy Coin]邮箱验证码", captcha, to, language)
		}()
	case MobileType:
		// TODO: Key添加个前缀，用于区别每个场景
		// 拼接Key
		mobileCaptcha := fmt.Sprintf("%s%s", model.MobileCaptcha, to)

		// 将验证码保存到Redis数据库中
		if _, err := RedisConn.Do("SET", mobileCaptcha, captcha, "EX", captchaMaxAge); err != nil {
			log.Errorf("sendCaptcha Error: %s", err.Error())
			return errors.New("服务器出错啦！")
		}

		go func() {
			sendMobileCaptcha(captcha, to, language)
		}()
	}
	return nil
}

type sendCaptchaByIDReq struct {
	UserId      uint64 `json:"user_id" binding:"required"`
	CaptchaType int32  `json:"captcha_type" binding:"required"`
	Language    string `json:"language"`
}

// 获取验证码
func SendCaptchaByID(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req sendCaptchaByIDReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	user, err := api.UserApi.GetUser(req.UserId)
	if err != nil {
		SendErrJSON("用户不存在", c)
		return
	}

	switch req.CaptchaType {
	case EmailType:
		if err := sendCaptcha(EmailType, user.Email, req.Language); err != nil {
			log.Errorf("SendCaptchaByID Error: %s", err.Error())
			SendErrJSON("发送邮箱验证码失败", c)
			return
		}
	case MobileType:
		if err := sendCaptcha(MobileType, user.Mobile, req.Language); err != nil {
			log.Errorf("SendCaptchaByID Error: %s", err.Error())
			SendErrJSON("发送手机验证码失败", c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{},
	})
}

// 校验是否是合法的手机号码
func isMobile(mobile string) bool {
	matched, _ := regexp.MatchString(`^(\w+)1[0-9]{10}$`, mobile)
	return matched
}

// 校验是否是合法的邮箱
func isEmail(email string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`, email)
	return matched
}

type sendCaptchaReq struct {
	Input       string 	`json:"input" binding:"required"`
	Type 		int32 	`json:"type" binding:"required"`
	Language    string 	`json:"language"`
}

// 发送验证码（注册）
func SendSignupCaptcha(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req sendCaptchaReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	exist, err := api.UserApi.IsExist(req.Input)
	if err != nil {
		SendErrJSON("服务器出错啦！", c)
		return
	}

	// 判断账户是否已经注册
	if exist {
		SendErrJSON(fmt.Sprintf("账户%s已被注册", req.Input), c)
		return
	}

	switch req.Type {
	case EmailType:
		if !isEmail(req.Input) {
			SendErrJSON("无效的邮箱账号", c)
			return
		}

		if err := sendCaptcha(EmailType, req.Input, req.Language); err != nil {
			log.Errorf("SendSignupCaptcha Error: %s", err.Error())
			SendErrJSON(err.Error(), c)
			return

		}
	case MobileType:
		if !isMobile(req.Input) {
			SendErrJSON("无效的手机号码", c)
			return
		}

		if err := sendCaptcha(MobileType, req.Input, req.Language); err != nil {
			log.Errorf("SendSignupCaptcha Error: %s", err.Error())
			SendErrJSON(err.Error(), c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{},
	})
}

// 发送验证码（找回密码）
func SendForgotCaptcha(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req sendCaptchaReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	exist, err := api.UserApi.IsExist(req.Input)
	if err != nil {
		SendErrJSON("服务器出错啦！", c)
		return
	}

	// 判断账户是否已经注册
	if !exist {
		SendErrJSON(fmt.Sprintf("账户%s尚未注册", req.Input), c)
		return
	}

	switch req.Type {
	case EmailType:
		if !isEmail(req.Input) {
			SendErrJSON("无效的邮箱账号", c)
			return
		}

		if err := sendCaptcha(EmailType, req.Input, req.Language); err != nil {
			log.Errorf("SendForgotCaptcha Error: %s", err.Error())
			SendErrJSON(err.Error(), c)
			return

		}
	case MobileType:
		if !isMobile(req.Input) {
			SendErrJSON("无效的手机号码", c)
			return
		}

		if err := sendCaptcha(MobileType, req.Input, req.Language); err != nil {
			log.Errorf("SendForgotCaptcha Error: %s", err.Error())
			SendErrJSON(err.Error(), c)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{},
	})
}

// 获取邮箱验证码（注册，带人机验证）
//func SendSignupEmailCaptchaForWeb(c *gin.Context) {
//	SendErrJSON := common.SendErrJSON
//
//	var req sendEmailCaptchaWebReq
//	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
//		SendErrJSON("参数无效", c)
//		return
//	}
//
//	if !isEmail(req.Email) {
//		SendErrJSON("无效的邮箱", c)
//		return
//	}
//
//	if req.LuosimaoCode == "" {
//		SendErrJSON("参数无效", c)
//		return
//	}
//
//	// 人机验证
//	verifyErr := utils.LuosimaoVerify(config.ServerConfig.LuosimaoVerifyURL, config.ServerConfig.LuosimaoAPIKey, req.LuosimaoCode)
//	if verifyErr != nil {
//		SendErrJSON(verifyErr.Error(), c)
//		return
//	}
//
//	// 判断需要绑定的该邮箱是否已经注册
//	if !model.DB.Model(model.User{}).Where("email = ?", req.Email).Find(&model.User{}).RecordNotFound() {
//		SendErrJSON("获取失败，邮箱已注册", c)
//		return
//	}
//
//	if err := sendCaptcha(EmailType, req.Email, req.Language); err != nil {
//		SendErrJSON(err.Error(), c)
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"errNo": model.ErrorCode.SUCCESS,
//		"msg":   "success",
//		"data": gin.H{
//		},
//	})
//}