package net

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"galaxyotc/common/model"
	"galaxyotc/common/log"
)

// Upload 文件上传
func Upload(c *gin.Context, uploadType int) (map[string]interface{}, error) {
	file, err := c.FormFile("galaxyFile")
	if err != nil {
		log.Errorf("Upload Error: %s", err.Error())
		return nil, errors.New("参数无效")
	}

	var filename = file.Filename
	var index = strings.LastIndex(filename, ".")

	if index < 0 {
		return nil, errors.New("无效的文件名")
	}

	var ext = filename[index:]
	if len(ext) == 1 {
		return nil, errors.New("无效的扩展名")
	}
	var mimeType = mime.TypeByExtension(ext)

	fmt.Printf("filename %s, index %d, ext %s, mimeType %s\n", filename, index, ext, mimeType)
	if mimeType == "" && ext == ".jpeg" {
		mimeType = "image/jpeg"
	}
	if mimeType == "" {
		return nil, errors.New("无效的图片类型")
	}

	var imgUploadedInfo model.ImageUploadedInfo

	switch uploadType {
	case Generate:
		imgUploadedInfo = model.ImgUploadedInfo(ext, "")
	case Currency:
		imgUploadedInfo = model.ImgUploadedInfo(ext, "currency")
	case TradingMethod:
		imgUploadedInfo = model.ImgUploadedInfo(ext, "trading_method")
	case UserAvatar:
		imgUploadedInfo = model.ImgUploadedInfo(ext, "avatar")
	}

	if err := os.MkdirAll(imgUploadedInfo.UploadDir, 0777); err != nil {
		log.Errorf("Upload Error: %s", err.Error())
		return nil, errors.New("error")
	}

	if err := c.SaveUploadedFile(file, imgUploadedInfo.UploadFilePath); err != nil {
		log.Errorf("Upload Error: %s", err.Error())
		return nil, errors.New("error1")
	}

	image := model.Image{
		Title:        imgUploadedInfo.Filename,
		OrignalTitle: filename,
		URL:          imgUploadedInfo.ImgURL,
		Width:        0,
		Height:       0,
		Mime:         mimeType,
	}

	return map[string]interface{}{
		"id":       image.ID,
		"url":      imgUploadedInfo.ImgURL,
		"title":    imgUploadedInfo.Filename, //新文件名
		"original": filename,                 //原始文件名
		"type":     mimeType,                 //文件类型
	}, nil
}

// UploadHandler 文件上传
func UploadHandler(c *gin.Context) {
	data, err := Upload(c, Generate)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errNo": model.ErrorCode.ERROR,
			"msg":   err.Error(),
			"data":  gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  data,
	})
}

// 币种图标上传
func CurrencyIconUploadHandler(c *gin.Context) {
	data, err := Upload(c, Currency)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errNo": model.ErrorCode.ERROR,
			"msg":   err.Error(),
			"data":  gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  data,
	})
}

// 交易方式图标上传
func TradingMethodIconUploadHandler(c *gin.Context) {
	data, err := Upload(c, TradingMethod)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errNo": model.ErrorCode.ERROR,
			"msg":   err.Error(),
			"data":  gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  data,
	})
}

func AvatarIconUploadHandler(c *gin.Context) {
	data, err := Upload(c, UserAvatar)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errNo": model.ErrorCode.ERROR,
			"msg":   err.Error(),
			"data":  gin.H{},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  data,
	})
}

const (
	Generate = 0
	Currency = 1
	TradingMethod = 2
	UserAvatar = 3
)