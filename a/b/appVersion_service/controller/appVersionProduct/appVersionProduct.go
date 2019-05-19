package appVersionProduct

import (
	"net/http"
	"galaxyotc/common/model"
	"github.com/gin-gonic/gin"
	"galaxyotc/common/log"
	"galaxyotc/common/net"
)

func AppVersionProduct(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	var appVeision model.AppVersion
	appProduct := c.Query("appType")
	if appProduct!="1" && appProduct!= "2"{
		SendErrJSON("参数不正确", c)
		return
	}

	if err := model.DB.Where("app_type = ? and enable = ?", appProduct,0).Order("-created_at").First(&appVeision).Error; err != nil {
		log.Errorf("User-Signin-Error: %s", err.Error())
		SendErrJSON("查询错误", c)
		return
	}


	c.JSON(http.StatusOK, gin.H {
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H {
			"appVeision": appVeision,
			},
	})
}
