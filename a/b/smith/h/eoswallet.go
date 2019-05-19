package gcwallet

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"strconv"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go"
)

// 导入EOS私钥
func (gcw *GalaxyCoinWallet) EosImportKey(key, configDir, password string) string {
	privateKey, err := ecc.NewPrivateKey(key)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	if err := gcw.eosWallet.ImportKey(privateKey, configDir, password); err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return fmt.Sprintf("{\"public_key\": %s}", privateKey.PublicKey().String())
}

// 返回所有已导入私钥的公钥和账户名信息
func (gcw *GalaxyCoinWallet) EosImportedKeys() string {
	keys, err := gcw.eosWallet.ImportedKeys()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	b, _ := json.Marshal(keys)
	return string(b)
}

// 切换当前的EOS账户
func (gcw *GalaxyCoinWallet) EosChangeAccount(key string) string {
	result := "false"
	publicKey, err := ecc.NewPublicKey(key)
	if err != nil {
		return result
	}
	if err := gcw.eosWallet.ChangeAccount(publicKey); err != nil {
		result = result
	}
	return "true"
}

// 返回所有的代币余额信息
func (gcw *GalaxyCoinWallet) EosTokenList() string {
	tl, err := gcw.eosWallet.GetTokenList()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	b, _ := json.Marshal(tl)
	return string(b)
}

// 生成一对新的私钥对
func (gcw *GalaxyCoinWallet) EosNewKeyPair() string {
	key, err := gcw.eosWallet.NewKeyPair()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	return fmt.Sprintf("{\"private_key\": %s, \"public_key\": \"%s\"}", key.String(), key.PublicKey().String())
}

// 获取当前的EOS账户
func (gcw *GalaxyCoinWallet) EosCurrentAccount() string {
	account := gcw.eosWallet.CurrentAccount()

	return string(*account)
}

// 当前EOS链高度
func (gcw *GalaxyCoinWallet) EosChainTip() string {

	height, hash := gcw.eosWallet.ChainTip()

	return fmt.Sprintf("{\"block_height\": %v, \"blockhash\": \"%s\"}", height, hash.String())
}

// 发送EOS
func (gcw *GalaxyCoinWallet) EosSpend(to, memo, amount string) string {
	hash, err := gcw.eosWallet.Spend(eos.AN(to), memo, amount)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return hash.String()
}

// 获取EOS账户余额
func (gcw *GalaxyCoinWallet) EosBalance() string {
	if string(*gcw.eosWallet.CurrentAccount()) == "" {
		return fmt.Sprintf("{\"confirmed\": %v, \"unconfirmed\": %v}", 0, 0)
	}
	balance := gcw.eosWallet.GetBalance(*gcw.eosWallet.CurrentAccount())
	return fmt.Sprintf("{\"confirmed\": %v, \"unconfirmed\": %v}", balance, 0)
}

type EosTxn struct {
	Txid 			string
	Height 			int32
	Timestamp 		Timestamp
	Value 			string
	Symbol 			string
	Status 			string
	Receiver		string
	Sender 			string
	Memo  			string
	Confirmations 	int32
}

type EosTxns struct {
	Txns 	[]*EosTxn		`json:"txns"`
	Count	int32			`json:"count"`
}

// 获取指定的EOS交易
func (gcw *GalaxyCoinWallet) EosGetTransaction(txid string) string {
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	txn, err := gcw.eosWallet.GetTransaction(*txHash)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	b, _ := json.Marshal(EosTxn{
		Txid:txn.Txid,
		Height:txn.Height,
		Timestamp:Timestamp(txn.Timestamp),
		Value:txn.Value,
		Symbol:txn.Symbol,
		Status: txn.Status,
		Receiver: txn.Receiver,
		Sender: txn.Sender,
		Memo: txn.Memo,
		Confirmations: txn.Confirmations,
	})
	return string(b)
}

// 获取EOS交易列表
func (gcw *GalaxyCoinWallet) EosGetTransactions(transactionType, sort, page, size int) string {
	// 校验查询交易类型参数是否合法
	switch transactionType {
	case 1, 2, 3:
	default:
		return fmt.Sprintf("error: %s", "transaction_type is invalid")
	}

	// 校验排序参数是否合法
	switch sort {
	case 1, 2:
	default:
		return fmt.Sprintf("error: %s", "sort is invalid")
	}

	txns, count, err := gcw.eosWallet.GetTransactions(transactionType, sort, page, size)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	var eosTxns EosTxns
	eosTxns.Count = count
	for _, txn:= range txns {
		eosTxns.Txns = append(eosTxns.Txns, &EosTxn{
			Txid:txn.Txid,
			Height:txn.Height,
			Timestamp:Timestamp(txn.Timestamp),
			Value:txn.Value,
			Symbol:txn.Symbol,
			Status: txn.Status,
			Receiver: txn.Receiver,
			Sender: txn.Sender,
			Memo: txn.Memo,
			Confirmations: txn.Confirmations,
		})
	}
	b, _ := json.Marshal(eosTxns)
	return string(b)
}

// 获取EOS资源信息
func (gcw *GalaxyCoinWallet) EosGetAccountResourceInfo() string {
	info, err := gcw.eosWallet.GetAccountResourceInfo()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	b, _ := json.Marshal(info)
	return string(b)
}

// 判断EOS金额是否为尘埃
func (gcw *GalaxyCoinWallet) EosIsDust(amount string) string {
	result := "false"
	if gcw.eosWallet.IsDust(amount) {
		result = "true"
	}
	return result
}

// 根据公钥查看是否已导入
func (gcw *GalaxyCoinWallet) EosHasKey(pk string) string {
	result := "false"
	publicKey, err := ecc.NewPublicKey(pk)
	if err != nil {
		return result
	}
	if gcw.eosWallet.HasKey(publicKey) {
		result = "true"
	}
	return result
}

// 获取账户信息
func (gcw *GalaxyCoinWallet) EosGetAccount() string {
	accountName := gcw.eosWallet.CurrentAccount()
	if string(*accountName) == "" {
		return fmt.Sprintln("please import the private key")
	}
	account, err := gcw.eosWallet.GetAccount(*accountName)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	b, _ := json.Marshal(account)
	return string(b)
}

// 获取汇率
func (gcw *GalaxyCoinWallet) EosExchangeRate(currencyCodeQuery string) string {
	exchangeRates := gcw.eosWallet.ExchangeRates()
	if exchangeRates == nil {
		return fmt.Sprintln("error: exchangeRates is nil")
	}
	rate, err := exchangeRates.GetExchangeRate(currencyCodeQuery)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return strconv.FormatFloat(rate, 'f', 18, 64)
}

// 获取一个新的随机账户名
func (gcw *GalaxyCoinWallet) EosNewAccountName() string {
	accountName, err := gcw.eosWallet.NewAccountName()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return string(accountName)
}

// 创建新账户
func (gcw *GalaxyCoinWallet) EosNewAccountByGC(key, name string) string {
	publicKey, err := ecc.NewPublicKey(key)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	var accountName eos.AccountName
	if name != "" {
		if len(name) != 12 {
			return fmt.Sprintf("error: %s", "account name length is must be 12")
		} else {
			accountName = eos.AN(name)
		}
	}
	hash, err := gcw.eosWallet.NewAccountByGC(publicKey, accountName)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	return hash.String()
}

// 导出密钥
func (gcw *GalaxyCoinWallet) EosExportKey(password string) string{
	privateKey, err := gcw.eosWallet.MasterPrivateKey(password)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	if privateKey == nil {
		return fmt.Sprintf("error: %s", "can not found private key, please import the private key first")
	}
	return fmt.Sprintf("{\"private_key\": %v, \"public_key\": %v}", privateKey.String(), privateKey.PublicKey().String())

}