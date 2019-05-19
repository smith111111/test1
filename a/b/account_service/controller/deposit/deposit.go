package deposit

import (
	"net/http"

	"galaxyotc/common/net"
	"galaxyotc/common/log"
	"galaxyotc/common/model"

	bwi "galaxyotc/btc-wallet-interface"
	wi "galaxyotc/wallet-interface"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"galaxyotc/gc_services/account_service/api"
	"time"
	"github.com/nats-io/go-nats"
	"galaxyotc/common/utils"

	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
)

// 获取指定币种的充值地址
func DepositAddress(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	code := c.Query("code")
	if code == "" {
		SendErrJSON("代币代码不能为空", c)
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(code)
	if err != nil {
		log.Errorf("Account-DepositAddress-Error: %s", err.Error())
		SendErrJSON("无效的代币类型", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	// 判断用户是否已经拥有未使用的充值地址，如果已经拥有则返回已经存在的地址
	var deposit model.Deposit
	notFound := model.DB.Where("account_id= ? and currency= ? and status= ?",user.ID, code, model.DepositNotInt).First(&deposit).RecordNotFound()
	if !notFound {
		c.JSON(http.StatusOK, gin.H{
			"errNo": model.ErrorCode.SUCCESS,
			"msg":   "success",
			"data": gin.H{
				"address": deposit.Address,
			},
		})
		return
	}

	// 获取充值地址
	switch currency.Family {
	case "ERC20":
		deposit.Address, err = api.EthereumWalletApi.Deposit(wi.EXTERNAL)
		if err != nil {
			log.Errorf("Account-DepositAddress-Error: %s", err.Error())
			SendErrJSON("获取充值地址有误", c)
			return
		}
	case "BTC":
		deposit.Address, err = api.MultiWalletApi.Deposit(currency.Code, currency.PropertyId, bwi.EXTERNAL)
		if err != nil {
			log.Errorf("Account-DepositAddress-Error: %s", err.Error())
			SendErrJSON("获取充值地址有误", c)
			return
		}
	case "EOS":
		deposit.Address = user.EosCode
	}

	deposit.Currency = currency.Code
	deposit.AccountID = user.ID
	deposit.Status = model.DepositNotInt

	// 避免重复，所以使用FirstOrCreate
	if err := model.DB.FirstOrCreate(&deposit, &model.Deposit{AccountID: deposit.AccountID, Currency: deposit.Currency, Address: deposit.Address}).Error; err != nil {
		log.Errorf("Account-DepositAddress-Error: %s", err.Error())
		SendErrJSON("获取充值地址有误", c)
		return
	}

	// 将待充值的地址存入Redis中
	if err := model.DepositToRedis(deposit); err != nil {
		log.Errorf("Account-DepositAddress-Error: %s", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"address": deposit.Address,
		},
	})
}

// 获取充值交易记录详情
func DepositDetail(c *gin.Context) {
	SendErrJSON := net.SendErrJSON

	sn := c.Param("sn")
	if sn == "" {
		SendErrJSON("交易流水号不能为空", c)
		return
	}

	// 从上下文管理器中获取用户信息
	userInter, _ := c.Get("user")
	user := userInter.(model.User)

	var depositInfo model.DepositInfo

	// 获取提币记录详情
	err := model.DB.Table("deposits").Where("sn = ? AND account_id = ?", sn, user.ID, ).First(&depositInfo).Error

	if err == gorm.ErrRecordNotFound {
		SendErrJSON("交易记录详情不存在", c)
		return
	} else if err != nil {
		log.Errorf("Account-DepositDetail-Error: %s", err.Error())
		SendErrJSON("服务器出错啦", c)
		return
	}

	// 根据状态码获取状态详情
	depositInfo.StatusString = model.DepositStatusDetail(depositInfo.Status)

	c.JSON(http.StatusOK, gin.H{
		"errNo": model.ErrorCode.SUCCESS,
		"msg":   "success",
		"data": gin.H{
			"depositInfo": depositInfo,
		},
	})
}

func UpdatePendingDeposit() {
	deposits := []*model.Deposit{}
	// 获取所有状态为充值中的记录信息
	if err := model.DB.Find(&deposits, model.Deposit{Status: model.DepositPendingInt}).Error; err != nil {
		log.Errorf("UpdatePendingDeposit-Error: %s", err.Error())
	}

	if len(deposits) > 0 {
		for _, deposit := range deposits {
			// 获取代币信息
			currency, err := model.CurrencyFromAndToRedis(deposit.Currency)
			if err != nil {
				log.Errorf("MultiCallback 获取代币信息失败: %s", err.Error())
				continue
			}

			// 获取钱包的当前高度
			currentHeight, _, err := api.MultiWalletApi.ChainTip(currency.Code, currency.PropertyId)
			if err != nil {
				log.Errorf("UpdatePendingDeposit-Error: %s", err.Error())
				continue
			}

			confirmations := int32(currentHeight) - deposit.Height + 1

			if confirmations >= 2 {
				doneAt := time.Now().Local()

				// 开启事务
				tx := model.DB.Begin()
				// 更改mysql数据库deposit表记录
				if err := tx.Model(&deposit).Updates(model.Deposit{Status: model.DepositCompletedInt, DoneAt: &doneAt, Confirmations: confirmations}).Error; err != nil {
					log.Errorf("multiDepositCallback 更新充值记录失败： %s", err.Error())
					tx.Rollback()
					continue
				}

				// 根据账户ID获取用户信息
				user, err := api.UserApi.GetUser(deposit.AccountID)
				if err != nil {
					log.Errorf("multiDepositCallback 获取用户信息失败： %s", err.Error())
					tx.Rollback()
					continue
				}

				amountWei := utils.ToWei(deposit.Amount, int(currency.Precision))

				// 调用私有钱包给用户挖等值的币
				if err := api.PrivateWalletApi.MintToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
					if err != nats.ErrTimeout {
						log.Errorf("multiDepositCallback 挖矿失败： %s", err.Error())
						continue
					}
				}
				tx.Commit()
			}
		}
	}
}