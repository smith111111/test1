package wallet

import (
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
	"github.com/spf13/viper"
)

// Transfer from tokens from an account to another
// transfer [from] [to] [amount]
func Transfer(from, to eos.AccountName, amount string) {
	contract := toAccount(viper.GetString("transfer-cmd-contract"), "--contract")

	var quantity eos.Asset
	var err error
	if contract == "eosio.token" {
		quantity, err = eos.NewEOSAssetFromString(amount)
	} else {
		quantity, err = eos.NewAsset(amount)
	}
	errorCheck("invalid amount", err)
	memo := viper.GetString("transfer-cmd-memo")

	api := getAPI()

	action := token.NewTransfer(from, to, quantity, memo)
	action.Account = contract
	pushEOSCActions(api, action)
}

