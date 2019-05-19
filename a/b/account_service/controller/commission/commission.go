package commission

import (
	"fmt"
	"errors"

	"galaxyotc/common/log"
	"galaxyotc/common/utils"
	"galaxyotc/common/model"

	"github.com/jinzhu/gorm"
	"galaxyotc/gc_services/account_service/api"

	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
)

// 发放佣金
func Distribution() (bool, error){
	startTime := utils.GetYesterdayTime()
	endTime := utils.GetTodayTime()

	distributionCommissionInfos := []*model.DistributionCommissionInfo{}
	// TODO: 过滤时间
	baseQuery := model.DB.Table("commission_distributions cd",
	).Select("user_id, currency, internal_address, private_token_address, cr.precision, sum(amount) as amount",
	).Joins("LEFT JOIN currencies cr ON cd.currency = cr.code",
	).Joins("LEFT JOIN users ur ON cd.user_id = ur.id",
	).Where("cd.status = ? and cd.created_at < ?", model.CommissionDistributionStatus_Draft, endTime,
	).Group("user_id, currency")

	if err := baseQuery.Find(&distributionCommissionInfos).Error; err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	// 发放
	for _, commissionInfo := range distributionCommissionInfos {
		// 1，执行发放
		log.Info(commissionInfo.PrivateTokenAddress+", "+commissionInfo.InternalAddress)
		success :=  transferToken(commissionInfo.PrivateTokenAddress, commissionInfo.InternalAddress, commissionInfo.Amount, commissionInfo.Precision)

		// 2，更新状态
		newStatus := model.CommissionDistributionStatus_Success
		if !success {
			newStatus = model.CommissionDistributionStatus_Fail
		}

		tx := model.DB.Begin()

		var commissionDistribution model.CommissionDistribution
		if err := model.DB.Model(&commissionDistribution).Where("user_id = ? and currency = ? and status = ? and created_at < ?" ,
			commissionInfo.UserID,
			commissionInfo.Currency,
			model.CommissionDistributionStatus_Draft,
			endTime).UpdateColumn("status", newStatus).Error; err != nil {
			tx.Rollback()
			log.Errorf("Commission Distribution 更新状态失败: %s", err.Error())
			continue
		}

		// 3，生成单据
		commissionDistributionReceipt := model.CommissionDistributionReceipt {
			StartAt: startTime,
			EndAt: endTime,
			Status: newStatus,
			UserID: commissionInfo.UserID,
			Currency: commissionInfo.Currency,
			Amount: commissionInfo.Amount,
		}
		if err := model.DB.Create(&commissionDistributionReceipt).Error; err != nil {
			tx.Rollback()
			log.Errorf("Commission Distribution 单据创建失败: %s", err.Error())
			continue
		}

		// TODO： 4， 记录水

		tx.Commit()
	}

	return true, nil
}

func transferToken(privateTokenAddress string, internalAddress string, amount float64, precision uint) bool {
	amountWei := utils.ToWei(amount, int(precision))

	if err := api.PrivateWalletApi.MintToken(privateTokenAddress, internalAddress, amountWei.String(), privateWalletApi.NoErrorCallback); err != nil {
		log.Errorf("Commission Distribution 佣金分配失败: %s", err.Error())
		return false
	}
	return true
}

// 提交佣金分发申请
func Apply(userID uint, businessType int32, orderID uint64, currency string, profit float64) (bool, error) {
	if orderID <= 0 {
		return false, errors.New("订单编号小于等于0")
	}
	if profit <= 0 {
		return false, errors.New("金额不能小于等于0")
	}
	if currency == "" {
		return false, errors.New("币种不能为空")
	}

	// 获取用户分佣比例
	var commissionRateList []*model.CommissionRate
	var err error
	if commissionRateList, err = getCommissionRate(userID); err != nil {
		return false, err
	}

	if len(commissionRateList) > 0 {
		// 添加佣金
		tx := model.DB.Begin()
		for _, commissionRateInfo := range commissionRateList {
			var commissionDistribution model.CommissionDistribution
			commissionDistribution.BusinessType = businessType
			commissionDistribution.OrderID = orderID
			commissionDistribution.UserID = commissionRateInfo.UserID
			commissionDistribution.UserType = commissionRateInfo.UserType
			commissionDistribution.Sn = ""
			commissionDistribution.Currency = currency
			commissionDistribution.Rate = commissionRateInfo.Rate
			commissionDistribution.Amount = commissionRateInfo.Rate * profit
			commissionDistribution.Status = model.CommissionDistributionStatus_Draft

			if err := tx.Create(&commissionDistribution).Error; err != nil {
				tx.Rollback()
				return false, err
			}
		}
		tx.Commit()
	}

	return true, nil
}

// 根据用户获取相关的返佣佣金比例
func getCommissionRate(userID uint) ([]*model.CommissionRate, error) {
	var commissionRateList []*model.CommissionRate
	var commissionUserList []*model.CommissionUser
	var err error
	if userID <= 0 {
		return  nil, errors.New("无效的用户")
	}

	var user model.User
	if err = model.DB.Table("users").Where("id = ?", userID).First(&user).Error; err != nil {
		fmt.Println(err.Error())
		return  nil, errors.New("获取用户信息失败")
	}

	if user.ParentId == 0 {
		return  nil, errors.New("没有分佣的会员")
	}
	if user.UserType == model.FoundingPartner {
		return  nil, errors.New("创世合伙人交易不分佣")
	}

	commissionUserList = getCommissionUsers(user.ParentId, 1)
	isAngelPartnerCommissioned := false
	var isFinish bool
	for _, commissionUser := range commissionUserList {
		var rate float64
		if commissionUser.UserType == model.RegularMembers {
			// 冻结用户不参与分佣，直接跳过
			if commissionUser.Level == 1 && commissionUser.Status == model.UserStatusNormal {
				commissionRateList = append(commissionRateList , &model.CommissionRate{UserID: commissionUser.UserID, UserType: commissionUser.UserType, Rate: model.CommissionRate_Parent})
			}
			continue

		} else if commissionUser.UserType == model.AngelPartner {
			if !isAngelPartnerCommissioned && commissionUser.Status == model.UserStatusNormal {
				// 天使合伙人：如果是直接上级，获得35%；否则为5%。
				if commissionUser.Level == 1 {
					rate = model.CommissionRate_Parent + 1 * model.CommissionRate_Partnership
				} else {
					rate = 1 * model.CommissionRate_Partnership
				}
				commissionRateList = append(commissionRateList , &model.CommissionRate{UserID: commissionUser.UserID, UserType: commissionUser.UserType, Rate: rate })
				isAngelPartnerCommissioned = true
			}
			continue

		} else if commissionUser.UserType == model.FoundingPartner {
			if commissionUser.Status == model.UserStatusNormal {
				// 创世合伙人：如果是直接上级，获得40%；第二级获得10%；否则5%。
				if commissionUser.Level == 1 {
					rate = model.CommissionRate_Parent + 2 * model.CommissionRate_Partnership
				} else if commissionUser.Level == 2	{
					if isAngelPartnerCommissioned {
						rate = model.CommissionRate_Partnership
					} else {
						rate = 2 * model.CommissionRate_Partnership
					}
				} else {
					rate = model.CommissionRate_Partnership
				}
				commissionRateList = append(commissionRateList , &model.CommissionRate{UserID: commissionUser.UserID, UserType: commissionUser.UserType, Rate: rate })
			}
			isFinish = true
		}

		if isFinish {
			break
		}
	}

	return commissionRateList, nil
}

// 获取所有分佣的用户
func getCommissionUsers(userID uint64, level int32) ([]*model.CommissionUser) {
	var commissionUsers []*model.CommissionUser
	var user model.User
	if err := model.DB.Table("users").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil
	}

	commissionUser := model.CommissionUser{UserID: user.ID, UserType: user.UserType, Status: user.Status, Level: level}
	commissionUsers = append(commissionUsers, &commissionUser)
	if user.ParentId == 0 {
		return commissionUsers
	}
	if nextcommissionUsers := getCommissionUsers(user.ParentId, level + 1); nextcommissionUsers != nil {
		commissionUsers = append(commissionUsers, nextcommissionUsers...)
	}

	return commissionUsers
}