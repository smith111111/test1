package gcwallet

import (
	"encoding/json"
	"fmt"
	ewi "gcwallet/eth-wallet-interface"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"strings"

)

// purpose: "internal", "external"
func (gcw *GalaxyCoinWallet) EthCurrentAddress(purpose string) string {
	var keyPurpose ewi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = ewi.INTERNAL
	case "external":
		keyPurpose = ewi.EXTERNAL
	default:
		keyPurpose = ewi.EXTERNAL
	}

	address := gcw.ethWallet.CurrentAddress(keyPurpose)

	return address.String()
}

// purpose: "internal", "external"
func (gcw *GalaxyCoinWallet) EthNewAddress(purpose string) string {
	var keyPurpose ewi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = ewi.INTERNAL
	case "external":
		keyPurpose = ewi.EXTERNAL
	default:
		keyPurpose = ewi.EXTERNAL
	}

	address := gcw.ethWallet.NewAddress(keyPurpose)

	return address.String()
}

func (gcw *GalaxyCoinWallet) EthChainTip() string {

	height, hash := gcw.ethWallet.ChainTip()

	return fmt.Sprintf("{\"block_height\": %v, \"blockhash\": \"%s\"}", height, hash.String())
}

// feeLevel: "economic", "normal", or "priority"
func (gcw *GalaxyCoinWallet) EthSpend(amount string, addr string, feeLevel string) string {
	ethAddr := common.HexToAddress(addr)
	address := ethAddress{&ethAddr}

	var fee_Level ewi.FeeLevel
	switch strings.ToLower(feeLevel) {
	case "economic":
		fee_Level = ewi.ECONOMIC
	case "normal":
		fee_Level = ewi.NORMAL
	case "priority":
		fee_Level = ewi.PRIOIRTY
	default:
		fee_Level = ewi.NORMAL
	}

	hash, err := gcw.ethWallet.Spend(amount,  address, fee_Level, "")
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return hash.String()
}

func (gcw *GalaxyCoinWallet) EthBalance() string {
	confirmed, unconfirmed := gcw.ethWallet.BalanceBigInt()
	return fmt.Sprintf("{\"confirmed\": %s, \"unconfirmed\": %s}", confirmed.String(), unconfirmed.String())
}

type EthTxn struct {
	Txid string
	Value string
	Height int32
	Timestamp Timestamp
	From string
	To string
	Gas string
	Confirmations int64
	Status string
	Input string
}

type EthTxns struct {
	Txns 	[]*EthTxn		`json:"txns"`
	Count	int32			`json:"count"`
}

func (gcw *GalaxyCoinWallet) EthGetTransaction(txid string) string {
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	txn, err := gcw.ethWallet.GetTransaction(*txHash)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	b, _ := json.Marshal(&EthTxn{
		Txid:txn.Txid.String(),
		Value:txn.Value,
		Height:txn.Height,
		Timestamp:Timestamp(txn.Timestamp),
		From: txn.From.String(),
		To: txn.To.String(),
		Gas: txn.Gas,
		Confirmations: txn.Confirmations,
		Status: string(txn.Status),
		Input: txn.Input,
	})
	return string(b)
}

func (gcw *GalaxyCoinWallet) EthTransactions(transactionType, sort, page, size int) string {
	// 计算起始位置
	offset := (page - 1) * size

	// 分页大小最多20条
	if size > 20 {
		size = 20
	}

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

	txns, count, err := gcw.ethWallet.Transactions(transactionType, sort, offset, size)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	var ethTxns EthTxns
	ethTxns.Count = count
	for _, txn:= range txns {
		ethTxns.Txns = append(ethTxns.Txns, &EthTxn{
			Txid:txn.Txid.String(),
			Value:txn.Value,
			Height:txn.Height,
			Timestamp:Timestamp(txn.Timestamp),
			From: txn.From.String(),
			To: txn.To.String(),
			Gas: txn.Gas,
			Confirmations: txn.Confirmations,
			Status: string(txn.Status),
			Input: txn.Input,
		})
	}
	b, _ := json.Marshal(ethTxns)
	return string(b)
}

// feeLevel: "economic", "normal", or "priority"
func (gcw *GalaxyCoinWallet) EthFeePerByte(feeLevel string) string {

	var fee_Level ewi.FeeLevel
	switch strings.ToLower(feeLevel) {
	case "economic":
		fee_Level = ewi.ECONOMIC
	case "normal":
		fee_Level = ewi.NORMAL
	case "priority":
		fee_Level = ewi.PRIOIRTY
	default:
		fee_Level = ewi.NORMAL
	}
	return fmt.Sprintf("{\"level\": %s, \"fee\": %v}", feeLevel, gcw.ethWallet.GetFeePerByte(fee_Level))
}

func (gcw *GalaxyCoinWallet) EthIsDust(amount string) string {
	amt, err := strconv.Atoi(amount)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	result := "false"
	if gcw.ethWallet.IsDust(int64(amt)) {
		result = "true"
	}
	return result
}

func (gcw *GalaxyCoinWallet) EthHasKey(addr string) string {
	ethAddr := common.HexToAddress(addr)
	address := ethAddress{&ethAddr}

	result := "false"
	if gcw.ethWallet.HasKey( address) {
		result = "true"
	}
	return result
}

func (gcw *GalaxyCoinWallet) EthExchangeRate(currencyCodeQuery string) string {
	exchangeRates := gcw.ethWallet.ExchangeRates()
	if exchangeRates == nil {
		return fmt.Sprintln("error: exchangeRates is nil")
	}
	rate, err := exchangeRates.GetExchangeRate(currencyCodeQuery)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return strconv.FormatFloat(rate, 'f', 18, 64)
}


// ZJ:
func (gcw *GalaxyCoinWallet) SuggestGasPrice() string {
	gasPrice, err := gcw.ethWallet.SuggestGasPrice()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return gasPrice.String()
}