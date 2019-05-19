package wallet

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/ryanuber/columnize"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
)

// retrieve the ABI associated with an account
func ABI(accountName eos.AccountName) {
	api := getAPI()

	abi, err := api.GetABI(accountName)
	errorCheck("get ABI", err)

	if !isStubABI(abi.ABI) {
		data, err := json.MarshalIndent(abi, "", "  ")
		errorCheck("json marshal", err)
		fmt.Println(string(data))
	} else {
		errorCheck("get abi", fmt.Errorf("no ABI has been set for account %q", accountName))
	}
}

// retrieve account information for a given name.  For a json dump, append the argument --json.
func Account(accountName eos.AccountName, printJson bool) {
	api := getAPI()

	account, err := api.GetAccount(accountName)
	errorCheck("get account", err)

	if printJson {
		data, err := json.MarshalIndent(account, "", "  ")
		errorCheck("json marshal", err)
		fmt.Println(string(data))
		return
	}
	printAccount(account)
}

func printAccount(account *eos.AccountResp) {
	if account != nil {
		// dereference this so we can safely mutate it to accomodate uninitialized symbols
		act := *account
		if act.SelfDelegatedBandwidth.CPUWeight.Symbol.Symbol == "" {
			act.SelfDelegatedBandwidth.CPUWeight.Symbol = act.TotalResources.CPUWeight.Symbol
		}
		if act.SelfDelegatedBandwidth.NetWeight.Symbol.Symbol == "" {
			act.SelfDelegatedBandwidth.NetWeight.Symbol = act.TotalResources.CPUWeight.Symbol
		}
		cfg := &columnize.Config{
			NoTrim: true,
		}

		for _, s := range []string{
			FormatBasicAccountInfo(&act, cfg),
			FormatPermissions(&act, cfg),
			FormatMemory(&act, cfg),
			FormatNetworkBandwidth(&act, cfg),
			FormatCPUBandwidth(&act, cfg),
			FormatBalances(&act, cfg),
			FormatProducers(&act, cfg),
			FormatVoterInfo(&act, cfg),
		} {
			fmt.Println(s)
			fmt.Println("")
		}
	}
}

// Retrieve currency balance for an account
func Balance(account, tokenContract eos.AccountName, symbol string) {
	api := getAPI()

	balances, err := api.GetCurrencyBalance(account, symbol, tokenContract)
	if err != nil {
		fmt.Printf("Error: get balance: %s\n", err)
		os.Exit(1)
	}

	for _, asset := range balances {
		fmt.Printf("%s\n", asset)
	}
}

// Get block data at a given height, or directly with a block hash
// block [block id | block height]
func Block(numOrIDRaw string) {
	api := getAPI()

	block, err := api.GetBlockByNumOrIDRaw(numOrIDRaw)
	errorCheck("get block", err)

	data, err := json.MarshalIndent(block, "", "  ")
	errorCheck("json marshaling", err)

	fmt.Println(string(data))
}

// retrieve the code associated with an account
func Code(accountName eos.AccountName) {
	api := getAPI()

	codeAndABI, err := api.GetRawCodeAndABI(accountName)
	errorCheck("get code", err)

	if codeAndABI.WASMasBase64 == "" {
		errorCheck("get code", fmt.Errorf("no code has been set for account %q", accountName))
		return
	}

	normalizedWASMBase64 := codeAndABI.WASMasBase64[:len(codeAndABI.WASMasBase64)-1]
	wasm, err := base64.StdEncoding.DecodeString(normalizedWASMBase64)
	errorCheck("decode WASM base64", err)

	hash := sha256.Sum256(wasm)
	fmt.Println("Code hash: ", hex.EncodeToString(hash[:]))

	if wasmFile := viper.GetString("get-code-cmd-output-wasm"); wasmFile != "" {
		err = ioutil.WriteFile(wasmFile, wasm, 0644)
		errorCheck("writing file", err)
		fmt.Printf("Wrote WASM to %q\n", wasmFile)
	}

	if abiFile := viper.GetString("get-code-cmd-output-raw-abi"); abiFile != "" {
		if codeAndABI.ABIasBase64 != "" {
			normalizedABIBase64 := codeAndABI.ABIasBase64[:len(codeAndABI.ABIasBase64)-1]

			abi, err := base64.StdEncoding.DecodeString(normalizedABIBase64)
			errorCheck("decode ABI base64", err)
			err = ioutil.WriteFile(abiFile, abi, 0644)
			errorCheck("writing file", err)
			fmt.Printf("Wrote ABI to %q\n", abiFile)
		} else {
			errorCheck("get code", fmt.Errorf("no ABI has been set for account %q", accountName))
		}
	}
}


// Retrieve blockchain infos, like head block, chain ID, etc..
func Info() {
	api := getAPI()

	info, err := api.GetInfo()
	errorCheck("get info", err)

	data, err := json.MarshalIndent(info, "", "  ")
	errorCheck("json marshal", err)

	fmt.Println(string(data))
}

// Get scheduled transactions pending for execution.
func ScheduledTransactions(lowerBound string, limit uint32) {
	api := getAPI()

	txs, err := api.GetScheduledTransactionsWithBounds(lowerBound, uint32(limit))
	errorCheck("get scheduled transactions", err)

	data, err := json.MarshalIndent(txs, "", "  ")
	errorCheck("json marshaling", err)

	fmt.Println(string(data))
}

// Fetch data from a table in a contract on chain
func Table(code,scope, table string) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:       code,
			Scope:      scope,
			Table:      table,
			LowerBound: viper.GetString("get-table-cmd-lower-bound"),
			UpperBound: viper.GetString("get-table-cmd-upper-bound"),
			Limit:      uint32(viper.GetInt("get-table-cmd-limit")),
			KeyType:    viper.GetString("get-table-cmd-key-type"),
			Index:      viper.GetString("get-table-cmd-index"),
			EncodeType: viper.GetString("get-table-cmd-encode-type"),
			JSON:       !(viper.GetBool("get-table-cmd-output-binary")),
		},
	)
	errorCheck("get table rows", err)

	data, err := json.MarshalIndent(response, "", "  ")
	errorCheck("json marshal", err)

	fmt.Println(string(data))
}