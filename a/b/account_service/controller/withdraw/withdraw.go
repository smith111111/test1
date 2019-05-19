package withdraw

import (
	"net/http"

	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"
	"galaxyotc/common/net"

	"github.com/rs/xid"
	"github.com/jinzhu/gorm"
	"github.com/gin-gonic/gin"
	"galaxyotc/gc_services/account_service/api"

	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
)

// 提币
func Withdraw(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	var withdraw model.Withdraw

	if err := c.ShouldBindJSON(&withdraw); err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
		SendErrJSON("参数无效", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(withdraw.Currency)
	if err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	/*if !user.IsRealName {
		SendErrJSON("请先进行实名认证", c)
		return
	}*/

	//RedisConn := model.RedisPool.Get()
	//defer RedisConn.Close()

	// 校验提币额度是否超过用户可提额度和当天已提额度
	//dayKey := model.WithdrawDayLimit + fmt.Sprintf("%d", user.ID)
	//dayCount, dayErr := redis.Float64(RedisConn.Do("GET", dayKey))
	//// 已提币额度加上本次额度等于当天总额度
	//if dayErr == nil && dayCount + withdraw.Amount >= ? {
	//	SendErrJSON("您今天的提币额度已超出限额。", c)
	//	return
	//}
	//
	//dayRemainingTime, _ := redis.Int64(RedisConn.Do("TTL", dayKey))
	//secondsOfDay := int64(24 * 60 * 60)
	//if dayRemainingTime < 0 || dayRemainingTime > secondsOfDay {
	//	dayRemainingTime = secondsOfDay
	//}
	//
	//if _, err := RedisConn.Do("SET", dayKey, dayCount + withdraw.Amount, "EX", dayRemainingTime); err != nil {
	//	fmt.Println("redis set failed:", err)
	//	SendErrJSON("内部错误.", c)
	//	return
	//}

	// 获取用户该代币的余额
	balanceWei, err := api.PrivateWalletApi.GetTokenBalance(currency.PrivateTokenAddress, user.InternalAddress)
	if err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
		SendErrJSON("获取账户余额失败", c)
		return
	}
	// 将余额由位转成小数
	blance, _ := utils.ToDecimal(balanceWei, int(currency.Precision)).Float64()

	if blance < withdraw.Amount {
		SendErrJSON("账户余额不足，无法进行提币", c)
		return
	}

	// 将用户传过来的小数转为位
	amountWei := utils.ToWei(withdraw.Amount, int(currency.Precision))
	// 获取对应币种的提币手续费并转成位
	FeeWei := utils.ToWei(currency.WithdrawFee, int(currency.Precision))
	// 实际提币金额等于提币金额减去手续费
	amountWei.Sub(amountWei, FeeWei)

	var txid string

	//// TODO 提币要进行审批，所以先进行锁币
	//switch currency.Family {
	//case "ERC20":
	//	if currency.Code == "ETH" {
	//		txid, err = api.EthereumWalletApi.EtherWithdraw(withdraw.Address, amountWei.String())
	//		if err != nil {
	//			log.Errorf("Account-Withdraw-Error: %s", err.Error())
	//			SendErrJSON(err.Error(), c)
	//			return
	//		}
	//	} else {
	//		txid, err = api.EthereumWalletApi.TokenWithdraw(currency.PublicTokenAddress, withdraw.Address, amountWei.String())
	//		if err != nil {
	//			log.Errorf("Account-Withdraw-Error: %s", err.Error())
	//			SendErrJSON(err.Error(), c)
	//			return
	//		}
	//	}
	//case "BTC":
	//	txid, err = api.MultiWalletApi.Withdraw(currency.Code, currency.PropertyId, withdraw.Address, amountWei.Int64(), bwi.NORMAL)
	//	if err != nil {
	//		log.Errorf("Account-Withdraw-Error: %s", err.Error())
	//		SendErrJSON(err.Error(), c)
	//		return
	//	}
	//case "EOS":
	//	txid, err = api.EosWalletApi.EosWithdraw(withdraw.Address, amountWei.String())
	//	if err != nil {
	//		log.Errorf("Account-Withdraw-Error: %s", err.Error())
	//		SendErrJSON(err.Error(), c)
	//		return
	//	}
	//}

	// 开启事务，数据保存与燃烧代币需要一致性
	tx := model.DB.Begin()
	// 将交易ID保存到数据库中，并修改提币状态为受理中
	withdraw.Sn = utils.BytesToHex([]byte(xid.New().String()))
	withdraw.Status = model.WithdrawPendingInt
	withdraw.AccountID = user.ID
	withdraw.Txid = txid
	withdraw.Fee = currency.WithdrawFee
	if err := tx.Create(&withdraw).Error; err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
		tx.Rollback()
		SendErrJSON("服务器出错啦！", c)
		return
	}

	// 由于返回的指针类型，所以之前转换的是扣除了手续费的金额，需要重新将用户传过来的小数转为位
	burnAmountWei := utils.ToWei(withdraw.Amount, int(currency.Precision))

	// 调用私有钱包扣除用户的提币金额
	if err := api.PrivateWalletApi.BurnToken(currency.PrivateTokenAddress, user.InternalAddress, burnAmountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
		tx.Rollback()
		SendErrJSON("服务器出错啦！", c)
		return
	}

	tx.Commit()

	// 将待完成的交易记录存入Redis中
	if err := model.WithdrawToRedis(withdraw); err != nil {
		log.Errorf("Account-Withdraw-Error: %s", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data":  gin.H{},
	})
}

// 获取提币交易记录详情
func WithdrawDetail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	sn := c.Param("sn")
	if sn == "" {
		SendErrJSON("交易流水号不能为空", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var withdrawInfo model.WithdrawInfo

	// 获取提币记录详情
	err := model.DB.Table("withdraws").Where("sn = ? AND account_id = ?", sn, user.ID).First(&withdrawInfo).Error

	if err == gorm.ErrRecordNotFound {
		SendErrJSON("交易记录详情不存在", c)
		return
	} else if err != nil {
		log.Errorf("Account-WithdrawDetail-Error: %s", err.Error())
		SendErrJSON("服务器出错啦", c)
		return
	}

	// 根据状态码获取状态详情
	withdrawInfo.StatusString = model.WithdrawStatusDetail(withdrawInfo.Status)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"withdrawInfo": withdrawInfo,
		},
	})
}
