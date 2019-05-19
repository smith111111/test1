package notice

import (
	"math"
	"net/http"
	"strconv"

	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/net"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func Notices(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	page, size := net.GetPageAndSize(c)
	offset := (page - 1) * size

	var (
		notices    []*model.Notice
		totalCount int64
	)
	noticeList := []*model.NoticeListInfo{}
	baseQuery := model.DB.Model(&model.Notice{}).Where("status = ?", model.Notice_ApproveSuccess).Order("created_at DESC")

	// 获取总数
	if err := baseQuery.Count(&totalCount).Error; err != nil {
		log.Errorf("Notices Error: %s", err.Error())
		SendErrJSON("获取公告列表失败", c)
		return
	}

	if totalCount > 0 {
		// 获取列表
		if err := baseQuery.Offset(offset).Limit(size).Find(&notices).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorf("Notices Error: %s", err.Error())
			SendErrJSON("获取公告列表失败", c)
			return
		}

		for _, notice := range notices {
			noticeListInfo := model.NoticeListInfo{
				ID:              notice.ID,
				Name:            notice.Name,
				Summary:         notice.Summary,
				CreatedAt:       notice.CreatedAt,
				CreatedAtString: notice.CreatedAt.Format("2006-01-02"),
			}
			noticeList = append(noticeList, &noticeListInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"notices":    noticeList,
			"pageNo":     page,
			"pageSize":   size,
			"totalPage":  math.Ceil(float64(totalCount) / float64(size)),
			"totalCount": totalCount,
		},
	})
}

func NoticeDetail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	noticeID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		SendErrJSON("无效的公告", c)
		return
	}

	var notice model.Notice
	if err := model.DB.First(&notice, noticeID).Error; err != nil {
		log.Errorf("NoticeDetail Error: %s", err.Error())
		SendErrJSON("无效的公告", c)
		return
	}

	if notice.Status != model.Notice_ApproveSuccess {
		SendErrJSON("无效的公告", c)
		return
	}

	noticeInfo := model.NoticeInfo{
		ID:        notice.ID,
		Name:      notice.Name,
		Summary:   notice.Summary,
		Content:   notice.Content,
		CreatedAt: notice.CreatedAt,
	}

	//TODO： format content

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"notice": noticeInfo,
		},
	})
}
