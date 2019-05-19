package gcwallet

import (
	"encoding/json"
	"fmt"
	wi "gcwallet/eth-wallet-interface"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"strings"
)

// purpose: "internal", "external"
func (gcw *GalaxyCoinWallet) GcCurrentAddress(purpose string) string {
	var keyPurpose wi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = wi.INTERNAL
	case "external":
		keyPurpose = wi.EXTERNAL
	default:
		keyPurpose = wi.EXTERNAL
	}

	address := gcw.gctWallet.CurrentAddress(keyPurpose)

	return address.String()
}

// purpose: "internal", "external"
func (gcw *GalaxyCoinWallet) GcNewAddress(purpose string) string {

	var keyPurpose wi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = wi.INTERNAL
	case "external":
		keyPurpose = wi.EXTERNAL
	default:
		keyPurpose = wi.EXTERNAL
	}

	address := gcw.gctWallet.NewAddress(keyPurpose)

	return address.String()
}

func (gcw *GalaxyCoinWallet) GcChainTip() string {

	height, hash := gcw.gctWallet.ChainTip()

	return fmt.Sprintf("{\"block_height\": %v, \"blockhash\": \"%s\"}", height, hash.String())
}

// feeLevel: "economic", "normal", or "priority"
func (gcw *GalaxyCoinWallet) GcSpend(amount string, addr string, feeLevel string) string {
	ethAddr := common.HexToAddress(addr)
	address := ethAddress{&ethAddr}

	var fee_Level wi.FeeLevel
	switch strings.ToLower(feeLevel) {
	case "economic":
		fee_Level = wi.ECONOMIC
	case "normal":
		fee_Level = wi.NORMAL
	case "priority":
		fee_Level = wi.PRIOIRTY
	default:
		fee_Level = wi.NORMAL
	}

	hash, err := gcw.gctWallet.Spend(amount, address, fee_Level, "")
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return hash.String()
}

func (gcw *GalaxyCoinWallet) GcBalance() string {
	confirmed, unconfirmed := gcw.gctWallet.BalanceBigInt()
	return fmt.Sprintf("{\"confirmed\": %s, \"unconfirmed\": %s}", confirmed.String(), unconfirmed.String())
}

func (gcw *GalaxyCoinWallet) GcGetTransaction(txid string) string {
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	txn, err := gcw.gctWallet.GetTransaction(*txHash)
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

func (gcw *GalaxyCoinWallet) GcTransactions(transactionType, sort, page, size int) string {
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

	txns, count, err := gcw.gctWallet.Transactions(transactionType, sort, offset, size)
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
func (gcw *GalaxyCoinWallet) GcFeePerByte(feeLevel string) string {

	var fee_Level wi.FeeLevel
	switch strings.ToLower(feeLevel) {
	case "economic":
		fee_Level = wi.ECONOMIC
	case "normal":
		fee_Level = wi.NORMAL
	case "priority":
		fee_Level = wi.PRIOIRTY
	default:
		fee_Level = wi.NORMAL
	}
	return fmt.Sprintf("{\"level\": %s, \"fee\": %v}", feeLevel, gcw.gctWallet.GetFeePerByte(fee_Level))
}

func (gcw *GalaxyCoinWallet) GcIsDust(amount string) string {
	amt, err := strconv.Atoi(amount)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	result := "false"
	if gcw.gctWallet.IsDust(int64(amt)) {
		result = "true"
	}
	return result
}

func (gcw *GalaxyCoinWallet) GcHasKey(addr string) string {
	ethAddr := common.HexToAddress(addr)
	address := ethAddress{&ethAddr}

	result := "false"
	if gcw.gctWallet.HasKey(address) {
		result = "true"
	}
	return result
}

func (gcw *GalaxyCoinWallet) GcExchangeRate(currencyCodeQuery string) string {
	exchangeRates := gcw.gctWallet.ExchangeRates()
	if exchangeRates == nil {
		return fmt.Sprintln("error: exchangeRates is nil")
	}
	rate, err := exchangeRates.GetExchangeRate(currencyCodeQuery)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return strconv.FormatFloat(rate, 'f', 18, 64)
}
