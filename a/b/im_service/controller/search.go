package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/gc_services/im_service/client"
	"galaxyotc/common/net"
)

type SearchUserInfoResp struct {
	Account    string `json:"account"`
	Avatar     string `json:"avatar"`
	CreateTime int32  `json:"createTime"`
	Custom     string `json:"custom"`
	Email      string `json:"email"`
	Gender     string `json:"gender"`
	Nick       string `json:"nick"`
	Tel        string `json:"tel"`
	UpdateTime int32  `json:"updateTime"`
}

// 搜索用户信息
func Search(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	keyword := c.Query("keyword")
	if keyword == "" {
		SendErrJSON("参数无效", c)
		return
	}

	pageNo, err := strconv.Atoi(c.Query("pageNo"))
	if err != nil {
		log.Errorf("Im-Search-Error: %s", err.Error())
		pageNo = 1
	}

	if pageNo < 1 {
		pageNo = 1
	}

	pageSize := 20
	offset := (pageNo - 1) * pageSize

	list, err := client.SearchServiceClient.SearchUserInfo(keyword, offset, pageSize)
	if err != nil {
		log.Errorf("Im-Search-Error: %s", err.Error())
		SendErrJSON("搜索用户信息失败", c)
		return
	}


	rsp := make([]*SearchUserInfoResp, len(list))
	for i, item := range list {
		v := &SearchUserInfoResp{
			Account: item.Code,
			Avatar: "",
			CreateTime: 0,
			Custom: "{}",
			Email: item.Email,
			Gender: "male",
			Nick: item.Name,
			Tel: item.Mobile,
			UpdateTime: 0,
		}
		rsp[i] = v
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"list": rsp,
		},
	})
}
