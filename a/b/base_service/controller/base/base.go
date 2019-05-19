package base

import (
	"net/http"

	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Version(c *gin.Context) {
	version := viper.GetString("server.version")

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"version": version,
		},
	})
}