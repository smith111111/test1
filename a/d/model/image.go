package model

import (
	"galaxyotc/common/utils"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"os"
	"strings"
	"unicode/utf8"
)

// Image 图片
type Image struct {
	ID           uint   `gorm:"primary_key" json:"id"`
	Title        string `json:"title"`
	OrignalTitle string `json:"orignalTitle"`
	URL          string `json:"url"`
	Width        uint   `json:"width"`
	Height       uint   `json:"height"`
	Mime         string `json:"mime"`
}

// ImageUploadedInfo 图片上传后的相关信息(目录、文件路径、文件名、UUIDName、请求URL)
type ImageUploadedInfo struct {
	UploadDir      string
	UploadFilePath string
	Filename       string
	UUIDName       string
	ImgURL         string
}

func ImgUploadedInfo(ext string, dirPath string) ImageUploadedInfo {
	sep := string(os.PathSeparator)
	uploadImgDir := viper.GetString("server.upload_img_dir")
	length := utf8.RuneCountInString(uploadImgDir)
	lastChar := uploadImgDir[length-1:]

	if dirPath == "" {
		dirPath = utils.GetTodayYM(sep)
	}

	var uploadDir string
	if lastChar != sep {
		uploadDir = uploadImgDir + sep + dirPath
	} else {
		uploadDir = uploadImgDir + dirPath
	}

	uuidName, _ := uuid.NewV4()
	filename := uuidName.String() + ext
	uploadFilePath := uploadDir + sep + filename
	imgURL := strings.Join([]string{
		//"http://" + viper.GetString("admin_service.img_host") + viper.GetString("admin_service.img_path"),
		viper.GetString("server.img_host") + viper.GetString("server.img_path"),
		dirPath,
		filename,
	}, "/")
	return ImageUploadedInfo{
		ImgURL:         imgURL,
		UUIDName:       uuidName.String(),
		Filename:       filename,
		UploadDir:      uploadDir,
		UploadFilePath: uploadFilePath,
	}
}
