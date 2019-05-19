package middleware

import (
	"galaxyotc/common/model"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// RefreshTokenCookie 刷新过期时间
func RefreshTokenCookie(c *gin.Context) {
	tokenString, err := c.Cookie("token")
	if tokenString != "" && err == nil {
		c.SetCookie("token", tokenString, viper.GetInt("server.token_max_age"), "/", "", true, true)
		if user, err := getUser(c); err == nil {
			model.UserToRedis(user)
		}
	}
	c.Next()
}
