package model

import (
	"github.com/spf13/viper"
	"github.com/garyburd/redigo/redis"
)

// 校验验证码
func VerifyCaptcha(captcha, captchaKey string) (bool, error) {
	// 开发/测试模式跳过验证
	if viper.GetBool("server.dev") && captcha == "AAAAAA" {
		return true, nil
	}

	RedisConn := RedisPool.Get()
	defer RedisConn.Close()

	// 获取Redis中的验证码，并判断是否一致
	redisCaptcha, err := redis.String(RedisConn.Do("GET", captchaKey))
	if err != nil {
		return false, err
	}

	if captcha != redisCaptcha {
		return false, err
	}

	// 将验证成功的验证码删掉
	_, _ = RedisConn.Do("DEL", captchaKey)

	return true, nil
}
