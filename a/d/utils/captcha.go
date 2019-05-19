package utils

import (
	"math/rand"
	"time"
)

const (
	NumberCaptcha = `0123456789`
	AlphaCaptcha = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	AlphaAndNumberCaptcha = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
)

// 指定位数的数字大小写字母验证码
func NewCaptcha(n int, captchaType string) string {
	var captchaBytes = []byte(captchaType)

	if n <= 0 {
		return ""
	}

	var bytes = make([]byte, n)
	var randBy bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		rand.Seed(time.Now().UnixNano())
		randBy = true
	}
	for i, b := range bytes {
		if randBy {
			bytes[i] = captchaBytes[rand.Intn(len(captchaBytes))]
		} else {
			bytes[i] = captchaBytes[b%byte(len(captchaBytes))]
		}
	}

	return string(bytes)
}