package middleware

import (
	"fmt"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"galaxyotc/common/model"
	"github.com/spf13/viper"
	"galaxyotc/common/net"
)

func getUser(c *gin.Context) (model.User, error) {
	var user model.User
	var tokenString string
	tokenString, _ = c.Cookie("token")
	// 如果Cookie中没有则尝试从Authorization中获取
	if tokenString == "" {
		authorizationString := c.Request.Header.Get("Authorization")
		if authorizationString == "" {
			return user, errors.New("未登录")
		}

		authorizationList := strings.Split(authorizationString, " ")
		if len(authorizationList) != 2 {
			return user, errors.New("未登录")
		}

		tokenString = authorizationList[1]
	}

	return CheckToken(tokenString)
}

func CheckToken(tokenStr string) (model.User, error) {
	var user model.User

	token, tokenErr := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("server.token_secret")), nil
	})

	if tokenErr != nil {
		return user, errors.New("未登录")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := uint64(claims["id"].(float64))
		var err error
		user, err = model.UserFromRedis(userID)
		if err != nil {
			return user, errors.New("未登录")
		}

		return user, nil
	}

	return user, errors.New("未登录")
}

// SetContextUser 给 context 设置 user
func SetContextUser(c *gin.Context) {
	var user model.User
	var err error
	if user, err = getUser(c); err != nil {
		c.Set("user", nil)
		c.Next()
		return
	}
	c.Set("user", user)
	c.Next()
}

// SigninRequired 必须是登录用户
func SigninRequired(c *gin.Context) {
	var user model.User
	var err error

	if user, err = getUser(c); err != nil {
		net.SendErrJSON("未登录", model.ErrorCode.LoginTimeout, c)
		return
	}

	c.Set("user", user)
	c.Next()
}
