package service

import (
	pb "galaxyotc/common/proto/backend/account"
	multiWalletPb "galaxyotc/common/proto/wallet/multi_wallet"
	ethereumWalletPb "galaxyotc/common/proto/wallet/ethereum_wallet"
	eosWalletPb "galaxyotc/common/proto/wallet/eos_wallet"
	privateWalletPb "galaxyotc/common/proto/wallet/private_wallet"

	"github.com/jinzhu/gorm"

	"galaxyotc/common/model"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/account_service/api"
	"github.com/golang/protobuf/proto"
	"time"
	"strconv"
	"math/big"
	"github.com/rs/xid"
	"galaxyotc/common/utils"
)

type db struct {
	*gorm.DB
}

func newDB() *db {
	return &db{model.DB}
}

// EthereumWallet 回调处理，判断是充值操作回调还是提币操作回调
func (db *db) EthereumCallback(req pb.CallbackReq) {
	// 反序列化
	var transaction ethereumWalletPb.TransactionCallback
	if err := proto.Unmarshal(req.Transaction, &transaction); err != nil {
		log.Errorf("EthereumCallback 解析参数失败: %s", err.Error())
		return
	}

	if transaction.IsDeposit {
		// 充值操作回调处理
		db.ethereumDepositCallback(transaction)
	} else {
		// 提币操作回调处理
		db.ethereumWithdrawCallback(transaction)
	}
}

// EosWallet 回调处理，判断是充值操作回调还是提币操作回调
func (db *db) EosCallback(req pb.CallbackReq) {
	// 反序列化
	var transaction eosWalletPb.TransactionCallback
	if err := proto.Unmarshal(req.Transaction, &transaction); err != nil {
		log.Errorf("EthereumCallback 解析参数失败: %s", err.Error())
		return
	}

	log.Infof("Transaction is: %+v", transaction)
	if transaction.IsDeposit {
		// 充值操作回调处理
		db.eosDepositCallback(transaction)
	} else {
		// 提币操作回调处理
		db.eosWithdrawCallback(transaction)
	}
}

// MultiWallet 回调处理，判断是充值操作回调还是提币操作回调
func (db *db) MultiCallback(req pb.CallbackReq) {
	// 反序列化
	var transaction multiWalletPb.TransactionCallback
	if err := proto.Unmarshal(req.Transaction, &transaction); err != nil {
		log.Errorf("MultiCallback 解析参数失败: %s", err.Error())
		return
	}

	log.Infof("Transaction is: %+v", transaction)
	var (
		isDeposit		bool
	)

	for _, out  := range transaction.Outputs {
		// 根据交易回调信息中的输出地址查询数据库中对应的充值信息
		if out.Address != "" {
			deposit, err := model.DepositFromRedis(out.Address)
			if err != nil {
				log.Debug("MultiCallback, Address:", out.Address)
				log.Infof("MultiCallback 查询充值记录失败： %s", err.Error())
				continue
			} else {
				isDeposit = true
				log.Debug("MultiCallback, Address:", out.Address, ", Currency:", deposit.Currency)
				// 将充值的记录修改为充值中，使用定时器不断监听直到达到指定的确认数修改为已完成
				db.UpdateMultiDepositPending(transaction, &deposit)
			}
		}
	}

	if !isDeposit {
		// 提币操作回调处理
		db.multiWithdrawCallback(transaction)
	}
}

func (db *db) UpdateMultiDepositPending(transaction multiWalletPb.TransactionCallback, deposit *model.Deposit) {
	// 获取代币信息
	currency, err := model.CurrencyFromAndToRedis(deposit.Currency)
	if err != nil {
		log.Errorf("UpdateDepositPending 获取代币信息失败: %s", err.Error())
		return
	}

	var height uint32
	if transaction.Height != 0 {
		height = uint32(transaction.Height)
	} else {
		// 获取钱包的当前高度
		currentHeight, _, err := api.MultiWalletApi.ChainTip(currency.Code, currency.PropertyId)
		if err != nil {
			log.Errorf("UpdateDepositPending 获取区块失败: %s", err.Error())
			return
		}

		height = currentHeight
	}

	var (
			amount    float64
			gas       float64
			txid      string
		)

	// 根据资产ID判断是否是USDT代币
	if currency.PropertyId != "0" {
		transaction, err := api.MultiWalletApi.GetOmniTransaction(currency.Code, transaction.Txid)
		if err != nil {
			log.Errorf("multiDepositCallback 获取Omni交易失败： %s", err.Error())
			return
		}
		amount, _ = strconv.ParseFloat(transaction.Amount, 64)
		gas, _ = strconv.ParseFloat(transaction.Fee, 64)
		txid = transaction.Txid
	} else {
		var inputValue int64
		// 遍历输入计算总输入
		for _, input := range transaction.Inputs {
			inputValue += input.Value
		}

		// 遍历输出计算总输出并根据地址取出充值数量
		var outputValue, totalAmount int64
		for _, output := range transaction.Outputs {
			outputValue += output.Value
			if output.Address == deposit.Address {
				totalAmount += output.Value
			}
		}

		// 矿工费等于总输入减去总输出
		gasWei := big.NewInt(inputValue - outputValue)
		gas, _ = utils.ToDecimal(gasWei, int(currency.Precision)).Float64()

		amountWei := big.NewInt(totalAmount)
		// 传回的数量是位的形式，所以转换成小数
		amount, _ = utils.ToDecimal(amountWei, int(currency.Precision)).Float64()

		txid = transaction.Txid
	}
	// 订单流水号
	sn := utils.BytesToHex([]byte(xid.New().String()))

	// 充值交易未达到指定确认数，将状态修改为充值中
	if err := db.Model(&deposit).Updates(model.Deposit{Sn: sn, Amount: amount, Gas: gas, Txid: txid, Status: model.DepositPendingInt, Confirmations: deposit.Confirmations, Height: int32(height)}).Error; err != nil {
		log.Errorf("UpdateDepositPending 修改充值记录状态失败: %s", err.Error())
		return
	}

	// 将已修改状态的充值的地址从Redis中删除
	if err := model.DepositDeleteRedis(deposit.Address); err != nil {
		log.Errorf("UpdateDepositPending Redis删除充值地址失败")
	}
}

// PrivateWallet 错误回调处理，判断是充值操作回调还是提币操作回调
func (db *db) PrivateErrorCallback(req pb.CallbackReq) {
	// 反序列化
	var transaction privateWalletPb.TransactionErrorCallback
	if err := proto.Unmarshal(req.Transaction, &transaction); err != nil {
		log.Errorf("PrivateErrorCallback 解析参数失败: %s", err.Error())
		return
	}

	log.Infof("Transaction is: %+v", transaction)
	if transaction.IsDeposit {
		db.depositErrorCallback(transaction)
	} else {
		db.withdrawErrorCallback(transaction)
	}
}

func (db *db) AccountTransfer(req pb.AccountTransferReq) {
	// 根据订单号获取转账交易记录
	var transfer model.Transfer
	if err := db.First(&transfer, model.Transfer{Sn: req.Sn, Status: model.TransferWaitingInt}).Error; err != nil {
		log.Errorf("AccountTransfer 获取转账交易记录失败: %s", err.Error())
		return
	}

	var status int32 = model.TransferSuccessInt
	// 执行多签名放币操作
	txid, err := api.PrivateWalletApi.ExecuteTransaction(req.Amount, req.Sn, 2, 0, req.ReceiverAddress, req.SenderAddress, req.TokenAddress, 0)
	if err != nil {
		log.Errorf("AccountTransfer 释放代币失败: %s", err.Error())
		// 放币失败将交易状态修改为失败
		status = model.TransferFailedInt
	}

	doneAt := time.Now().Local()
	if err := model.DB.Model(&transfer).Updates(model.Transfer{Txid: txid, Status: status, DoneAt: &doneAt}).Error; err != nil {
		log.Errorf("AccountTransfer 保存转账交易记录: %s", err.Error())
	}
}