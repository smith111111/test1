package utils

import (
	"regexp"
	"strings"
)

// 显示前三后四
func TextReplace(text string) string {
	var reg *regexp.Regexp
	textLen := len(text)
	if textLen <= 6 {
		return text
	} else if textLen == 7 {
		reg = regexp.MustCompile("(\\d{3})\\d{0}(\\d{4})")
	} else if textLen == 8 {
		reg = regexp.MustCompile("(\\d{3})\\d{1}(\\d{4})")
	} else if textLen == 9 {
		reg = regexp.MustCompile("(\\d{3})\\d{2}(\\d{4})")
	} else if textLen == 10 {
		reg = regexp.MustCompile("(\\d{3})\\d{3}(\\d{4})")
	} else if textLen >= 11 {
		reg = regexp.MustCompile("(\\d{3})\\d{4}(\\d{4})")
	}
	return reg.ReplaceAllString(text, "$1****$2")
}

func TextReplaceForEmail(email string) string {
	prefix := strings.Split(email, "@")[0]
	if len(prefix) > 2 {
		prefix = string([]byte(prefix)[:2])
	} else {
		prefix = ""
	}

	reg := regexp.MustCompile(`(?:\w{2}).*?(?:\w@)`)
	return prefix + reg.ReplaceAllString(email, "****@")
}

// 校验手机格式
func RegexpMobile(mobile string) bool {
	matched, err := regexp.MatchString(`^1[0-9]{10}$`, mobile)
	if err != nil {
		return false
	}
	return matched
}

// 校验邮箱格式
func RegexpEmail(email string) bool {
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`, email)
	if err != nil {
		return false
	}
	return matched
}