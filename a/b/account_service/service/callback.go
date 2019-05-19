package service

import (
	"time"
	"strconv"
	"math/big"

	multiWalletPb "galaxyotc/common/proto/wallet/multi_wallet"
	ethereumWalletPb "galaxyotc/common/proto/wallet/ethereum_wallet"
	eosWalletPb "galaxyotc/common/proto/wallet/eos_wallet"
	privateWalletPb "galaxyotc/common/proto/wallet/private_wallet"
	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"

	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/utils"

	"github.com/rs/xid"
	"galaxyotc/gc_services/account_service/api"
	"github.com/nats-io/go-nats"
)

// EosWallet 充值操作回调处理
func (db *db) eosDepositCallback(transaction eosWalletPb.TransactionCallback) {
	// 根据交易回调地址查询数据库中对应的充值信息
	deposit, err := model.DepositFromRedis(transaction.Memo)
	if err != nil {
		log.Infof("eosDepositCallback 查询充值记录失败： %s", err.Error())
		return
	}

	var currency model.Currency
	if err := model.DB.Where("code = ?", "EOS").First(&currency).Error; err != nil {
		log.Errorf("eosDepositCallback 查询代币信息失败： %s", err.Error())
		return
	}

	if deposit.Currency != currency.Code {
		log.Error("eosDepositCallback 收到的交易与充值记录中的代币类型不符")
		return
	}

	// 转换交易金额
	amount, _ := utils.ToDecimal(transaction.Quantity, int(currency.Precision)).Float64()

	// 订单流水号
	sn := utils.BytesToHex([]byte(xid.New().String()))

	// 开启事务
	tx := db.Begin()

	doneAt := time.Unix(transaction.BlockTime, 0)
	// 更改mysql数据库deposit表记录
	if err := tx.Model(&deposit).Updates(model.Deposit{Sn: sn, Amount: amount, Gas: 0, Txid: transaction.Txid, Status: model.DepositCompletedInt, DoneAt: &doneAt, Confirmations: 1}).Error; err != nil {
		log.Errorf("eosDepositCallback 更新充值记录失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 将已经完成充值的地址从Redis中删除
	if err := model.DepositDeleteRedis(deposit.Address); err != nil {
		log.Errorf("eosDepositCallback Redis删除充值地址失败")
	}

	// 根据账户ID获取用户信息
	user, err := api.UserApi.GetUser(deposit.AccountID)
	if err != nil {
		log.Errorf("eosDepositCallback 获取用户信息失败： %s", err.Error())
		tx.Rollback()
		return
	}

	amountWei, _ := new(big.Int).SetString(transaction.Quantity, 10)

	// 调用私有钱包给用户挖等值的币
	if err := api.PrivateWalletApi.MintToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		if err != nats.ErrTimeout {
			log.Errorf("eosDepositCallback 挖矿失败： %s", err.Error())
		}
	}

	tx.Commit()

	// 充值成功后给推送消息给用户
}

// EosWallet 提币操作回调处理
func (db *db) eosWithdrawCallback(transaction eosWalletPb.TransactionCallback) {
	// 根据交易ID获取对应的提币记录
	withdraw, err := model.WithdrawFromRedis(transaction.Txid)
	if err != nil {
		log.Infof("eosWithdrawCallback 查询提币记录失败： %s", err.Error())
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(withdraw.Currency)
	if err != nil {
		log.Errorf("eosWithdrawCallback 获取代币信息失败")
		return
	}

	// 开启事务
	tx := db.Begin()

	doneAt := time.Unix(transaction.BlockTime, 0)
	// 更新对应的提币记录信息
	if err := tx.Model(&withdraw).Updates(model.Withdraw{Gas: 0, Status: model.WithdrawCompletedInt, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("eosWithdrawCallback 更新提币记录失败： %s", err.Error())
		return
	}

	// 将已经完成更新的提币记录从Redis中删除
	if err := model.WithdrawDeleteRedis(withdraw.Txid); err != nil {
		log.Errorf("eosWithdrawCallback Redis删除提币记录失败")
	}

	// 根据账户ID获取用户信息
	user, err := api.UserApi.GetUser(withdraw.AccountID)
	if err != nil {
		log.Errorf("eosWithdrawCallback 获取用户信息失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 交易信息中的金额是扣除手续费后的金额，所以这里应该用数据库中的金额
	amountWei := utils.ToWei(withdraw.Amount, int(currency.Precision))

	// 调用私有钱包给用户烧掉等值的币
	if err := api.PrivateWalletApi.BurnToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		if err != nats.ErrTimeout {
			log.Errorf("eosWithdrawCallback 烧币失败： %s", err.Error())
		}
	}

	tx.Commit()

	// 提币成功后给推送消息给用户
}

// EthereumWallet 充值操作回调处理
func (db *db) ethereumDepositCallback(transaction ethereumWalletPb.TransactionCallback) {
	// 根据交易回调地址查询数据库中对应的充值信息
	deposit, err := model.DepositFromRedis(transaction.To)
	if err != nil {
		log.Infof("ethereumDepositCallback 查询充值记录失败： %s", err.Error())
		return
	}

	var currency model.Currency
	// 智能合约地址为空则是ETH类型
	if transaction.Contract == "" {
		if err := model.DB.Where("code = ?", "ETH").First(&currency).Error; err != nil {
			log.Errorf("ethereumDepositCallback 查询代币信息失败： %s", err.Error())
			return
		}
	} else {
		if err := model.DB.Where("public_token_address = ?", transaction.Contract).First(&currency).Error; err != nil {
			log.Errorf("ethereumDepositCallback 查询代币信息失败： %s", err.Error())
			return
		}
	}

	if deposit.Currency != currency.Code {
		log.Error("ethereumDepositCallback 收到的交易与充值记录中的代币类型不符")
		return
	}

	// 转换交易金额
	amount, _ := utils.ToDecimal(transaction.Value, int(currency.Precision)).Float64()

	// 计算矿工费
	gasPrice, _ := new(big.Int).SetString(transaction.GasPrice, 10)
	gasInt := gasPrice.Mul(gasPrice, big.NewInt(int64(transaction.Gas)))
	gas, _ := utils.ToDecimal(gasInt, int(currency.Precision)).Float64()

	// 订单流水号
	sn := utils.BytesToHex([]byte(xid.New().String()))

	// 开启事务
	tx := db.Begin()

	doneAt := time.Unix(transaction.BlockTime, 0)
	// 更改mysql数据库deposit表记录
	if err := tx.Model(&deposit).Updates(model.Deposit{Sn: sn, Amount: amount, Gas: gas, Txid: transaction.Txid, Status: model.DepositCompletedInt, DoneAt: &doneAt, Confirmations: 1}).Error; err != nil {
		log.Errorf("ethereumDepositCallback 更新充值记录失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 将已经完成充值的地址从Redis中删除
	if err := model.DepositDeleteRedis(deposit.Address); err != nil {
		log.Errorf("ethereumDepositCallback Redis删除充值地址失败")
	}

	// 根据账户ID获取用户信息
	user, err := api.UserApi.GetUser(deposit.AccountID)
	if err != nil {
		log.Errorf("ethereumDepositCallback 获取用户信息失败： %s", err.Error())
		tx.Rollback()
		return
	}

	amountWei, _ := new(big.Int).SetString(transaction.Value, 10)

	// 调用私有钱包给用户挖等值的币
	if err := api.PrivateWalletApi.MintToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		if err != nats.ErrTimeout {
			log.Errorf("ethereumDepositCallback 挖矿失败： %s", err.Error())
		}
	}
	tx.Commit()

	// 充值成功后给推送消息给用户
}

// EthereumWallet 提币操作回调处理
func (db *db) ethereumWithdrawCallback(transaction ethereumWalletPb.TransactionCallback) {
	// 根据交易ID获取对应的提币记录
	withdraw, err := model.WithdrawFromRedis(transaction.Txid)
	if err != nil {
		log.Infof("ethereumWithdrawCallback 查询提币记录失败： %s", err.Error())
		return
	}

	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(withdraw.Currency)
	if err != nil {
		log.Errorf("ethereumWithdrawCallback 获取代币信息失败")
		return
	}

	// 计算矿工费
	gasPrice, _ := new(big.Int).SetString(transaction.GasPrice, 10)
	gasInt := gasPrice.Mul(gasPrice, big.NewInt(int64(transaction.Gas)))
	gas, _ := utils.ToDecimal(gasInt, int(currency.Precision)).Float64()

	// 开启事务
	tx := db.Begin()

	doneAt := time.Unix(transaction.BlockTime, 0)
	// 更新对应的提币记录信息
	if err := tx.Model(&withdraw).Updates(model.Withdraw{Gas: gas, Status: model.WithdrawCompletedInt, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("ethereumWithdrawCallback 更新提币记录失败： %s", err.Error())
		return
	}

	// 将已经完成更新的提币记录从Redis中删除
	if err := model.WithdrawDeleteRedis(withdraw.Txid); err != nil {
		log.Errorf("ethereumWithdrawCallback Redis删除提币记录失败")
	}

	// 根据账户ID获取用户信息
	user, err := api.UserApi.GetUser(withdraw.AccountID)
	if err != nil {
		log.Errorf("ethereumWithdrawCallback 获取用户信息失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 交易信息中的金额是扣除手续费后的金额，所以这里应该用数据库中的金额
	amountWei := utils.ToWei(withdraw.Amount, int(currency.Precision))

	// 调用私有钱包给用户烧掉等值的币
	if err := api.PrivateWalletApi.BurnToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		if err != nats.ErrTimeout {
			log.Errorf("ethereumWithdrawCallback 烧币失败： %s", err.Error())
		}
	}
	tx.Commit()

	// 提币成功后给推送消息给用户
}

// MultiWallet 充值操作回调处理
//func (db *db) multiDepositCallback(transaction multiWalletPb.TransactionCallback, deposit *model.Deposit, currency *model.Currency) {
//	// 获取交易金额
//	var (
//		amount    float64
//		amountWei *big.Int
//		gas       float64
//		txid      string
//		doneAt    time.Time
//	)
//
//	// 根据资产ID判断是否是USDT代币
//	if currency.PropertyId != "0" {
//		transaction, err := api.MultiWalletApi.GetOmniTransaction(currency.Code, transaction.Txid)
//		if err != nil {
//			log.Errorf("multiDepositCallback 获取Omni交易失败： %s", err.Error())
//			return
//		}
//		amount, _ = strconv.ParseFloat(transaction.Amount, 64)
//		// 传回的数量是小数的形式，所以转换成位
//		amountWei = utils.ToWei(amount, int(currency.Precision))
//		gas, _ = strconv.ParseFloat(transaction.Fee, 64)
//		txid = transaction.Txid
//		// Omni交易返回的是时间戳需要转换
//		doneAt = transaction.BlockTime
//	} else {
//		var inputValue int64
//		// 遍历输入计算总输入
//		for _, input := range transaction.Inputs {
//			inputValue += input.Value
//		}
//
//		// 遍历输出计算总输出并根据地址取出充值数量
//		var outputValue, totalAmount int64
//		for _, output := range transaction.Outputs {
//			outputValue += output.Value
//			if output.Address == deposit.Address {
//				totalAmount += output.Value
//			}
//		}
//
//		// 矿工费等于总输入减去总输出
//		gasWei := big.NewInt(inputValue - outputValue)
//		gas, _ = utils.ToDecimal(gasWei, int(currency.Precision)).Float64()
//
//		amountWei = big.NewInt(totalAmount)
//		// 传回的数量是位的形式，所以转换成小数
//		amount, _ = utils.ToDecimal(amountWei, int(currency.Precision)).Float64()
//
//		txid = transaction.Txid
//		doneAt = time.Unix(transaction.Timestamp, 0)
//	}
//	// 订单流水号
//	sn := utils.BytesToHex([]byte(xid.New().String()))
//
//	// 开启事务
//	tx := db.Begin()
//	// 更改mysql数据库deposit表记录
//	if err := tx.Model(&deposit).Updates(model.Deposit{Sn: sn, Amount: amount, Gas: gas, Txid: txid, Status: model.DepositCompletedInt, DoneAt: &doneAt, Confirmations: deposit.Confirmations}).Error; err != nil {
//		log.Errorf("multiDepositCallback 更新充值记录失败： %s", err.Error())
//		tx.Rollback()
//		return
//	}
//
//	// 将已经完成充值的地址从Redis中删除
//	if err := model.DepositDeleteRedis(deposit.Address); err != nil {
//		log.Errorf("multiDepositCallback Redis删除充值地址失败")
//	}
//
//	// 根据账户ID获取用户信息
//	user, err := api.UserApi.GetUser(deposit.AccountID)
//	if err != nil {
//		log.Errorf("multiDepositCallback 获取用户信息失败： %s", err.Error())
//		tx.Rollback()
//		return
//	}
//
//	// 调用私有钱包给用户挖等值的币
//	if err := api.PrivateWalletApi.MintToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String()); err != nil {
//		if err != nats.ErrTimeout {
//			log.Errorf("multiDepositCallback 挖矿失败： %s", err.Error())
//		}
//	}
//	tx.Commit()
//
//	// TODO 充值成功后给推送消息给用户
//}

// MultiWallet 提币操作回调处理
func (db *db) multiWithdrawCallback(transaction multiWalletPb.TransactionCallback) {
	// 根据交易ID获取对应的提币记录
	withdraw, err := model.WithdrawFromRedis(transaction.Txid)
	if err != nil {
		log.Infof("multiWithdrawCallback 查询提币记录失败： %s", err.Error())
		return
	}

	// 获取货币信息
	currency, err := model.CurrencyFromAndToRedis(withdraw.Currency)
	if err != nil {
		log.Error("multiWithdrawCallback 获取货币信息失败: %s", err.Error())
		return
	}

	if transaction.Height <= 0 {
		return
	}

	// 获取钱包的当前高度
	currentHeight, _, err := api.MultiWalletApi.ChainTip(currency.Code, currency.PropertyId)
	if err != nil {
		log.Errorf("multiWithdrawCallback 获取区块失败: %s", err.Error())
		return
	}
	// 计算确认数,确认数等于当前高度减去交易高度
	height := int32(currentHeight) - transaction.Height + 1

	// TODO 指定交易确认数
	if height < 2 {
		return
	}

	// 获取交易金额
	var (
		gas    float64
		doneAt time.Time
	)

	// 根据资产ID判断是否是USDT代币
	if currency.PropertyId != "0" {
		transaction, err := api.MultiWalletApi.GetOmniTransaction(currency.Code, transaction.Txid)
		if err != nil {
			log.Errorf("multiDepositCallback 获取Omni交易失败： %s", err.Error())
			return
		}

		gas, _ = strconv.ParseFloat(transaction.Fee, 64)
		// Omni交易返回的是时间戳需要转换
		doneAt = transaction.BlockTime
	} else {
		var inputValue int64
		// 遍历输入计算总输入
		for _, input := range transaction.Inputs {
			inputValue += input.Value
		}

		// 遍历输出计算总输出并根据地址取出提币数量
		var outputValue, totalAmount int64
		for _, output := range transaction.Outputs {
			outputValue += output.Value
			if output.Address == withdraw.Address {
				totalAmount += output.Value
			}
		}

		// 矿工费等于总输入减去总输出
		gasWei := big.NewInt(inputValue - outputValue)
		gas, _ = utils.ToDecimal(gasWei, int(currency.Precision)).Float64()

		doneAt = time.Unix(transaction.Timestamp, 0)
	}

	// 开启事务
	tx := db.Begin()

	// 更新对应的提币记录信息
	if err := tx.Model(&withdraw).Updates(model.Withdraw{Gas: gas, Status: model.WithdrawCompletedInt, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("multiWithdrawCallback 更新提币记录失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 将已经完成更新的提币记录从Redis中删除
	if err := model.WithdrawDeleteRedis(withdraw.Txid); err != nil {
		log.Errorf("multiWithdrawCallback Redis删除提币记录失败: %s", err.Error())
	}

	// 根据账户ID获取用户信息
	user, err := api.UserApi.GetUser(withdraw.AccountID)
	if err != nil {
		log.Errorf("multiWithdrawCallback 获取用户信息失败： %s", err.Error())
		tx.Rollback()
		return
	}

	// 交易信息中的金额是扣除手续费后的金额，所以这里应该用数据库中的金额
	amountWei := utils.ToWei(withdraw.Amount, int(currency.Precision))

	// 调用私有钱包给用户烧掉等值的币
	if err := api.PrivateWalletApi.BurnToken(currency.PrivateTokenAddress, user.InternalAddress, amountWei.String(), privateWalletApi.TokenErrorCallback); err != nil {
		if err != nats.ErrTimeout {
			log.Errorf("multiWithdrawCallback 烧币失败： %s", err.Error())
		}
	}

	tx.Commit()

	// TODO 提币成功后给推送消息给用户
}

// PrivateWallet 充值操作错误回调处理
func (db *db) depositErrorCallback(transaction privateWalletPb.TransactionErrorCallback) {
	// 根据用户内部地址获取用户信息
	user, err := api.UserApi.GetUserByInternalAddress(transaction.WhoAddress)
	if err != nil {
		log.Errorf("depositErrorCallback 获取用户信息失败： %s", err.Error())
		return
	}

	// 根据用户ID和充值地址获取对应的充值记录
	var deposit model.Deposit
	if err := model.DB.Where("address = ? AND account_id = ?", transaction.WhoAddress, user.Id).First(&deposit).Error; err != nil {
		log.Errorf("depositErrorCallback 获取充值记录失败： %s, Token address is: %s, User ID is: %d", err.Error(), transaction.TokenAddress, user.Id)
		return
	}

	// 更改充值状态为异常
	if err := model.DB.Model(&deposit).Updates(model.Deposit{Status: model.DepositAbnormalInt}).Error; err != nil {
		log.Errorf("depositErrorCallback 更新充值记录异常状态失败： %s", err.Error())
	}
}

// PrivateWallet 提币操作错误回调处理
func (db *db) withdrawErrorCallback(transaction privateWalletPb.TransactionErrorCallback) {
	// 根据用户内部地址获取用户信息
	user, err := api.UserApi.GetUserByInternalAddress(transaction.WhoAddress)
	if err != nil {
		log.Errorf("withdrawErrorCallback 获取用户信息失败： %s", err.Error())
		return
	}

	// 根据用户ID和充值地址获取对应的充值记录
	var withdraw model.Withdraw
	if err := model.DB.Where("address = ? AND account_id = ?", transaction.WhoAddress, user.Id).First(&withdraw).Error; err != nil {
		log.Errorf("withdrawErrorCallback 获取提币记录失败： %s", err.Error())
		return
	}

	// 更改充值状态为异常
	if err := model.DB.Model(&withdraw).Updates(model.Withdraw{Status: model.WithdrawAbnormalInt}).Error; err != nil {
		log.Errorf("withdrawErrorCallback 更新提币记录异常状态失败： %s", err.Error())
	}
}
