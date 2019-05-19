package area

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/model"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/mozillazg/go-pinyin"
)

type newAreaReq struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	Sort       int    `json:"sort"`
}

// 创建新区域码
func NewArea(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req newAreaReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	var newArea model.Area
	if !model.DB.Where("name = ? OR code = ?", req.Name, req.Code).NewRecord(&newArea) {
		SendErrJSON("区域名或区域代码已存在", c)
		return
	}

	newArea.Name = req.Name
	newArea.Code = req.Code
	newArea.Status = model.AreaStatusEnable
	newArea.Sort = req.Sort

	// 简写拼音，只取首字母
	pinyinArgs := pinyin.NewArgs()
	pinyinArgs.Style = pinyin.FirstLetter
	newArea.PinYin = strings.Join(pinyin.LazyPinyin(req.Name, pinyinArgs),"")

	// 全拼
	newArea.FullPinYin = strings.Join(pinyin.LazyConvert(req.Name, nil), "")

	if err := model.DB.Create(&newArea).Error; err != nil {
		SendErrJSON("创建区域码失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// Delete 删除一个区域
func Delete(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var id int
	var idErr error
	if id, idErr = strconv.Atoi(c.Param("id")); idErr != nil {
		SendErrJSON("无效的id", c)
		return
	}

	var area model.Area

	if err := model.DB.First(&area, id).Error; err != nil {
		SendErrJSON("无效的id", c)
		return
	}

	if err :=  model.DB.Delete(&area).Error; err != nil {
		SendErrJSON("error", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"id": id,
		},
	})
}

type updateAreaReq struct {
	ID         int    	`json:"id"`
	Status     int 		`json:"status"`
	Sort       int    	`json:"sort"`
	PinYin     string 	`json:"pinyin"`
	FullPinYin string 	`json:"fullpinyin"`
}

// UpdateArea 更新区域
func UpdateArea(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var req updateAreaReq
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		SendErrJSON("参数无效", c)
		return
	}

	var updateArea model.Area
	if err := model.DB.First(&updateArea, req.ID).Error; err != nil {
		log.Errorf("UpdateArea Error: %s", err.Error())
		SendErrJSON("无效的区域码", c)
		return
	}

	updateArea.Sort = req.Sort
	updateArea.Status = req.Status
	updateArea.PinYin = req.PinYin
	updateArea.FullPinYin = req.FullPinYin

	if err := model.DB.Save(&updateArea).Error; err != nil {
		log.Errorf("UpdateArea Error: %s", err.Error())
		SendErrJSON("更新区域码失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 根据筛选条件获取区域列表
func Areas(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	// 获取页数和条数
	page, size := net.GetPageAndSize(c)
	// 计算起始位置
	offset := (page - 1) * size

	var (
		selectSQL string
		args      []interface{}
	)

	var (
		areas      []*model.Area
		totalCount int64
	)

	baseQuery := model.DB.Model(model.Area{}).Where(selectSQL, args...)

	if err := baseQuery.Offset(offset).Limit(size).Find(&areas).Order("sort DESC").Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("Areas Error: %s", err.Error())
		SendErrJSON("获取区域列表失败", c)
		return
	}

	// 获取区域总数量
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Areas Error: %s", err.Error())
		SendErrJSON("获取区域列表失败", c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"areas":     areas,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}
