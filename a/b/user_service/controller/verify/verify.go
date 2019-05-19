package verify

import (
	"time"
	"errors"
	"net/http"
	"encoding/json"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type IdcardResult struct {
	ErrorCode int          `json:"error_code"`
	Reason    string       `json:"reason"`
	Result    PersonResult `json:"result"`
}
type PersonResult struct {
	RalName string `json:"realname"`
	IdCard  string `json:"idcard"`
	IsOk    bool   `json:"isok"`
}

//实名认证
func Verify(c *gin.Context) {
	SendErrJSON := net.SendErrJSON
	type VerifyIDCardRequestModel struct {
		CardType int32   `json:"card_type"`
		RealName string `json:"name"`
		CardNo   string `json:"card_no"`
		AreaCode string `json:"area_code"`
	}

	var requestData VerifyIDCardRequestModel
	if err := c.ShouldBindWith(&requestData, binding.JSON); err != nil {
		log.Errorf("User-Verify-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	if user.IsRealName {
		SendErrJSON("你已通过实名认证", c)
		return
	}

	// 验证该用户是否已提交过申请
	var realnameVerification model.RealnameVerification
	if err := model.DB.Where("user_id = ?", user.ID).Last(&realnameVerification).Error; err == nil {
		if realnameVerification.Status == utils.Apply || realnameVerification.Status == utils.Approvaling {
			SendErrJSON("你已申请过实名认证，请等待审批", c)
			return
		} else if realnameVerification.Status == utils.Approvaled {
			SendErrJSON("你已通过实名认证", c)
			return
		}
	}

	// 验证证件是否已被其他用户使用
	if err := model.DB.Where("id_card_type =? and area_code = ? and id_card_no = ?", requestData.CardType, requestData.AreaCode, requestData.CardNo).Find(&realnameVerification).Error; err == nil {
		if realnameVerification.UserId != user.ID {
			SendErrJSON("该证件已被使用", c)
			return
		}
	}

	//内地身份证直接通过第三方服务进行认证
	if requestData.CardType == model.CardType_ID && requestData.AreaCode == "86" {
		if succeed, err := verifyIDCard_ZH(requestData.RealName, requestData.CardNo); err != nil || !succeed {
			log.Errorf("User-Verify-Error: %s", err.Error())
			SendErrJSON("身份证号码与当前姓名不符合", c)
			return
		}

		// 开始事务
		tx := model.DB.Begin()

		if err := tx.Model(&user).Updates(model.User{IsRealName: true, RealnameVerificationStatus: model.RealnameVerificationStatus_Approved}).Error; err != nil {
			tx.Rollback()
			log.Errorf("User-Verify-Error: %s", err.Error())
			SendErrJSON("服务器错误", c)
			return
		}

		var newRealnameVerification model.RealnameVerification
		newRealnameVerification.AreaCode = requestData.AreaCode
		newRealnameVerification.IDCardType = requestData.CardType
		newRealnameVerification.IDCardNo = requestData.CardNo
		newRealnameVerification.IDCardName = requestData.RealName
		newRealnameVerification.Status = utils.Approvaled
		newRealnameVerification.ApplyTime = time.Now().Local()
		timeNow := time.Now().Local()
		newRealnameVerification.ApprovedTime = &timeNow
		newRealnameVerification.ApprovedUser = "admin"
		newRealnameVerification.UserId = user.ID

		if err := tx.Create(&newRealnameVerification).Error; err != nil {
			tx.Rollback()
			log.Errorf("User-Verify-Error: %s", err.Error())
			SendErrJSON("服务器错误", c)
			return
		}

		// 提交事务
		tx.Commit()
	} else {
		// 开始事务
		tx := model.DB.Begin()

		if err := tx.Model(&user).Update("realname_verification_status", model.RealnameVerificationStatus_Verifying).Error; err != nil {
			tx.Rollback()
			log.Errorf("User-Verify-Error: %s", err.Error())
			SendErrJSON("服务器错误", c)
			return
		}

		var newRealnameVerification model.RealnameVerification
		newRealnameVerification.AreaCode = requestData.AreaCode
		newRealnameVerification.IDCardType = requestData.CardType
		newRealnameVerification.IDCardNo = requestData.CardNo
		newRealnameVerification.IDCardName = requestData.RealName
		newRealnameVerification.Status = utils.Apply
		newRealnameVerification.ApplyTime = time.Now().Local()
		newRealnameVerification.UserId = user.ID
		if err := model.DB.Create(&newRealnameVerification).Error; err != nil {
			tx.Rollback()
			log.Errorf("User-Verify-Error: %s", err.Error())
			SendErrJSON("服务器错误", c)
			return
		}

		// 提交事务
		tx.Commit()
	}

	if err := model.UserToRedis(user); err != nil {
		log.Errorf("更新Redis数据失败: %s", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"is_realname": user.IsRealName,
		},
	})
}

//验证内地身份证
func verifyIDCard_ZH(realName, cardNo string) (bool, error) {
	var url = "http://aliyunverifyidcard.haoservice.com/idcard/VerifyIdcardv2"
	var param = "realName=" + realName + "&cardNo=" + cardNo
	var resp = utils.RequestGet(url, param)

	var result IdcardResult
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return false, errors.New("身份证验证失败")
	}
	return result.Result.IsOk, nil
}
