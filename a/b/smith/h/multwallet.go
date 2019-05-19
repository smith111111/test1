package gcwallet

import (
	"encoding/json"
	"fmt"
	wi "gcwallet/btc-wallet-interface"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"strconv"
	"strings"
)

// purpose: "internal", "external"
// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiCurrentAddress(currencyCode, purpose string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	var keyPurpose wi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = wi.INTERNAL
	case "external":
		keyPurpose = wi.EXTERNAL
	default:
		keyPurpose = wi.EXTERNAL
	}

	address := btcWallet.CurrentAddress(keyPurpose)

	return address.String()
}

// purpose: "internal", "external"
// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiNewAddress(currencyCode, purpose string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	var keyPurpose wi.KeyPurpose
	switch strings.ToLower(purpose) {
	case "internal":
		keyPurpose = wi.INTERNAL
	case "external":
		keyPurpose = wi.EXTERNAL
	default:
		keyPurpose = wi.EXTERNAL
	}

	address := btcWallet.NewAddress(keyPurpose)

	return address.String()
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiChainTip(currencyCode string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	height, hash := btcWallet.ChainTip()

	return fmt.Sprintf("{\"block_height\": %v, \"blockhash\": \"%s\"}", height, hash.String())
}

// feeLevel: "economic", "normal", or "priority"
// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiSpend(currencyCode string, amount string, addr string, feeLevel string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	address, err := btcWallet.DecodeAddress(addr)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

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
	amt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	hash, err := btcWallet.Spend(amt, address, fee_Level, "")
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return hash.String()
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiBalance(currencyCode string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	confirmed, unconfirmed := btcWallet.Balance()
	return fmt.Sprintf("{\"confirmed\": %v, \"unconfirmed\": %v}", confirmed, unconfirmed)
}


type MultiTxn struct {
	Txid string
	Value int64
	Height int32
	Timestamp Timestamp
	WatchOnly bool
	Confirmations int64
	Status string
	ErrorMessage string
	Bytes []byte
}

type MultiTxns struct {
	Txns 	[]*MultiTxn		`json:"txns"`
	Count	int32			`json:"count"`
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiGetTransaction(currencyCode, txid string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	txn, err := btcWallet.GetTransaction(*txHash)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	b, _ := json.Marshal(MultiTxn{
		Txid:txn.Txid,
		Value:txn.Value,
		Height:txn.Height,
		Timestamp:Timestamp(txn.Timestamp),
		WatchOnly:txn.WatchOnly,
		Confirmations:txn.Confirmations,
		Status:string(txn.Status),
		ErrorMessage:txn.ErrorMessage,
		Bytes:txn.Bytes,
	})
	return string(b)
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiTransactions(currencyCode string, transactionType, sort, page, size int) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	// 计算起始位置
	offset := (page - 1) * size

	// 分页大小最多20条
	if size > 20 {
		size = 20
	}

	// 校验排序参数是否合法
	switch sort {
	case 1, 2:
	default:
		return fmt.Sprintf("error: %s", "sort is invalid")
	}

	txns, count, err := btcWallet.Transactions(transactionType, sort, offset, size)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	var multiTxns MultiTxns
	multiTxns.Count = count
	for _, txn:= range txns {
		multiTxns.Txns = append(multiTxns.Txns, &MultiTxn{
			Txid:txn.Txid,
			Value:txn.Value,
			Height:txn.Height,
			Timestamp:Timestamp(txn.Timestamp),
			WatchOnly:txn.WatchOnly,
			Confirmations:txn.Confirmations,
			Status:string(txn.Status),
			ErrorMessage:txn.ErrorMessage,
			Bytes:txn.Bytes,
		})
	}
	b, _ := json.Marshal(multiTxns)
	return string(b)
}

// feeLevel: "economic", "normal", or "priority"
// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiFeePerByte(currencyCode, feeLevel string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
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
	return fmt.Sprintf("{\"level\": \"%s\", \"fee\": %v}", feeLevel, btcWallet.GetFeePerByte(fee_Level))
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiIsDust(currencyCode, amount string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	amt, err := strconv.Atoi(amount)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	result := "false"
	if btcWallet.IsDust(int64(amt)) {
		result = "true"
	}
	return result
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
func (gcw *GalaxyCoinWallet) MultiHasKey(currencyCode, addr string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	address, err := btcWallet.DecodeAddress(addr)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	result := "false"
	if btcWallet.HasKey(address) {
		result = "true"
	}
	return result
}

// currencyCode: "BTC", "BCH", "ZEC", "LTC", "TBTC", "TBCH", "TZEC", "TLTC"
// currencyCode: "USD", "BTC", "CNY"...
func (gcw *GalaxyCoinWallet) MultiExchangeRate(currencyCode, currencyCodeQuery string) string {
	btcWallet, err := gcw.multiWallet.WalletForCurrencyCode(currencyCode)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	exchangeRates := btcWallet.ExchangeRates()
	if exchangeRates == nil {
		return fmt.Sprintln("error: exchangeRates is nil")
	}
	rate, err := exchangeRates.GetExchangeRate(currencyCodeQuery)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return strconv.FormatFloat(rate, 'f', 8, 64)
}
