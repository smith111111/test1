package net

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"galaxyotc/common/model"
	"strconv"
)

// 返回错误信息
func SendErrJSON(msg string, args ...interface{}) {
	if len(args) == 0 {
		panic("缺少 *gin.Context")
	}
	var c *gin.Context
	var errNo = model.ErrorCode.ERROR
	if len(args) == 1 {
		theCtx, ok := args[0].(*gin.Context)
		if !ok {
			panic("缺少 *gin.Context")
		}
		c = theCtx
	} else if len(args) == 2 {
		theErrNo, ok := args[0].(int)
		if !ok {
			panic("errNo不正确")
		}
		errNo = theErrNo
		theCtx, ok := args[1].(*gin.Context)
		if !ok {
			panic("缺少 *gin.Context")
		}
		c = theCtx
	}
	c.JSON(http.StatusOK, gin.H{
		"errNo": errNo,
		"msg":   msg,
		"data":  gin.H{},
	})
	// 终止请求链
	c.Abort()
}

// 获取分页页数和分页条数
func GetPageAndSize(c *gin.Context) (int32, int32) {
	var err error
	// 获取分页页数,默认第一页
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
		err = nil
	}

	if page < 1 {
		page = 1
	}

	// 获取分页条数, 默认一页十条
	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		size = 10
		err = nil
	}

	if size < 1 {
		size = 10
	}

	return int32(page), int32(size)
}
