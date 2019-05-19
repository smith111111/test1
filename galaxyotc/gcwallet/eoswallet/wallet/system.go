package wallet

import (
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/spf13/viper"
	"os"
)

// Bid on a premium account name.
// bidname [bidder_account_name] [premium_account_name] [bid quantity]
func Bidname(bidder, newName eos.AccountName, quantity string) {
	api := getAPI()

	bidAsset, err := eos.NewEOSAssetFromString(quantity)
	errorCheck("bid amount invalid", err)

	fmt.Printf("[%s] bidding for: %s , amount=%d precision=%d symbol=%s\n", bidder, newName, bidAsset.Amount, bidAsset.Symbol.Precision, bidAsset.Symbol.Symbol)

	pushEOSCActions(api,
		system.NewBidname(bidder, newName, bidAsset),
	)
}

// Buy RAM at market price, for a given number of bytes.
func BuyRamBytes(payer, receiver eos.AccountName, numBytes int64) {
	api := getAPI()

	if int64(uint32(numBytes)) != numBytes {
		fmt.Printf("Invalid number of bytes: capped at unsigned 32 bits.  That's probably too much RAM anyway.\n")
		os.Exit(1)
	}

	pushEOSCActions(api,
		system.NewBuyRAMBytes(payer, receiver, uint32(numBytes)),
	)
}

// Cancel a deferred transaction. Use 'eosc tx cancel' instead
func CancelDelay() {
	fmt.Println("Use `eosc tx cancel` instead.")
}

// Claim block production rewards. Once per day, don't forget it!
func ClaimRewards(owner eos.AccountName) {
	api := getAPI()

	pushEOSCActions(api,
		system.NewClaimRewards(owner),
	)
}

// Delegate some CPU and Network bandwidth, to yourself or others.
// delegatebw [from] [receiver] [network bw stake qty] [cpu bw stake qty]
// The --transfer option makes it so the receiver will be able to unstake
//what was delegated to them, and receive the corresponding EOS back. It
//is effectively transfering the coins to them.
func DelegateBW(from, receiver eos.AccountName, net_stake, cpu_stake string, transfer bool) {
	netStake, err := eos.NewEOSAssetFromString(net_stake)
	errorCheck(`"network bw stake qty" invalid`, err)
	cpuStake, err := eos.NewEOSAssetFromString(cpu_stake)
	errorCheck(`"cpu bw stake qty" invalid`, err)

	api := getAPI()

	pushEOSCActions(api, system.NewDelegateBW(from, receiver, cpuStake, netStake, transfer))
}

// Removes a permission currently set on an account. See --help for more details.
// deleteauth [account] [permission_name]
func DeleteAuth(account eos.AccountName, permissionName eos.Name) {
	api := getAPI()
	pushEOSCActions(api, system.NewDeleteAuth(account, eos.PermissionName(permissionName)))
}

// Assign a permission to the given code::action pair.
// linkauth [your account] [code account] [action name] [permission name]
func LinkAuth(account, code eos.AccountName, actionName eos.ActionName, permission eos.PermissionName) {

	api := getAPI()

	pushEOSCActions(api, system.NewLinkAuth(account, code, actionName, permission))
}

// Create a new account.
// newaccount [creator] [new_account_name]
func NewAccount(creator, newAccount eos.AccountName) {
	var actions []*eos.Action
	authFile := viper.GetString("system-newaccount-cmd-auth-file")
	authKey := viper.GetString("system-newaccount-cmd-auth-key")
	if authKey == "" && authFile == "" {
		fmt.Println("Error: pass one of --auth-file or --auth-key")
		os.Exit(1)
	}

	if authKey != "" && authFile != "" {
		fmt.Println("Error: pass either --auth-file or --auth-key")
		os.Exit(1)
	}

	if authFile != "" {
		// load from YAML
		var authStruct struct {
			Owner  eos.Authority `json:"owner"`
			Active eos.Authority `json:"active"`
		}
		err := loadYAMLOrJSONFile(authFile, &authStruct)
		errorCheck("auth-file invalid", err)

		if authStruct.Owner.Threshold == 0 {
			errorCheck("auth-file invalid", fmt.Errorf("owner struct missing?"))
		}

		if authStruct.Active.Threshold == 0 {
			errorCheck("auth-file invalid", fmt.Errorf("active struct missing?"))
		}

		actions = append(actions, system.NewCustomNewAccount(creator, newAccount, authStruct.Owner, authStruct.Active))
	} else {
		// authKey then
		pubKey, err := ecc.NewPublicKey(authKey)
		errorCheck("parsing public key", err)

		actions = append(actions, system.NewNewAccount(creator, newAccount, pubKey))
	}

	cpuStakeStr := viper.GetString("system-newaccount-cmd-stake-cpu")
	netStakeStr := viper.GetString("system-newaccount-cmd-stake-net")

	if cpuStakeStr == "" {
		errorCheck("missing argument", fmt.Errorf("--stake-cpu missing"))
	}
	if netStakeStr == "" {
		errorCheck("missing argument", fmt.Errorf("--stake-net missing"))
	}

	cpuStake, err := eos.NewEOSAssetFromString(cpuStakeStr)
	errorCheck("--stake-cpu invalid", err)
	netStake, err := eos.NewEOSAssetFromString(netStakeStr)
	errorCheck("--stake-net invalid", err)

	doTransfer := viper.GetBool("system-newaccount-cmd-transfer")
	actions = append(actions, system.NewDelegateBW(creator, newAccount, cpuStake, netStake, doTransfer))

	buyRAM := viper.GetString("system-newaccount-cmd-buy-ram")
	if buyRAM != "" {
		buyRAMAmount, err := eos.NewEOSAssetFromString(buyRAM)
		errorCheck("--buy-ram invalid", err)

		actions = append(actions, system.NewBuyRAM(creator, newAccount, uint64(buyRAMAmount.Amount)))
	} else {
		buyRAMBytes := viper.GetInt("system-newaccount-cmd-buy-ram-kbytes")
		actions = append(actions, system.NewBuyRAMBytes(creator, newAccount, uint32(buyRAMBytes*1024)))
	}

	if viper.GetBool("system-newaccount-cmd-setpriv") {
		actions = append(actions, system.NewSetPriv(newAccount))
	}

	api := getAPI()

	pushEOSCActions(api, actions...)
}

// Register an account as a block producer candidate.
// regproducer [account_name] [public_key] [website_url]
func RegProducer(accountName eos.AccountName, public_key, website_url string) {
	api := getAPI()

	publicKey, err := ecc.NewPublicKey(public_key)
	errorCheck(fmt.Sprintf("%q invalid public key", public_key), err)
	websiteURL := website_url

	pushEOSCActions(api,
		system.NewRegProducer(accountName, publicKey, websiteURL, uint16(viper.GetInt("system-regproducer-cmd-location"))),
	)
}

// Register an account as a voting proxy.
func RegProxy(accountName eos.AccountName) {
	api := getAPI()

	pushEOSCActions(api,
		system.NewRegProxy(accountName, true),
	)
}

// Sell the [num bytes] amount of bytes of RAM on the RAM market.
func SellRam(accountName eos.AccountName, numBytes int64) {
	api := getAPI()

	pushEOSCActions(api,
		system.NewSellRAM(accountName, uint64(numBytes)),
	)
}

// Set ABI only on an account.
// setabi [account name] [abi file]
func SetAbi(accountName eos.AccountName,abiFile string) {
	api := getAPI()

	action, err := system.NewSetABI(accountName, abiFile)
	errorCheck("loading abi file", err)

	pushEOSCActions(api, action)
}

// Set code only on an account.
// setcode [account name] [wasm file]
func SetCode(accountName eos.AccountName, wasmFile string) {
	api := getAPI()

	action, err := system.NewSetCode(accountName, wasmFile)
	errorCheck("loading wasm file", err)

	pushEOSCActions(api, action)
}

// Set code only on an account.
// setcode [account name] [wasm file]
func SetContract(accountName eos.AccountName, wasmFile, abiFile string) {
	api := getAPI()

	actions, err := system.NewSetContract(accountName, wasmFile, abiFile)
	errorCheck("loading files", err)

	pushEOSCActions(api,
		actions...,
	)
}

// Undelegate some CPU and Network bandwidth.
// undelegatebw [from] [receiver] [network bw unstake qty] [cpu bw unstake qty]
func UndelegateBW(from, receiver eos.AccountName, net_stake, cpu_stake string) {
	netStake, err := eos.NewEOSAssetFromString(net_stake)
	errorCheck(`"network bw unstake qty" invalid`, err)
	cpuStake, err := eos.NewEOSAssetFromString(cpu_stake)
	errorCheck(`"cpu bw unstake qty" invalid`, err)

	api := getAPI()

	pushEOSCActions(api, system.NewUndelegateBW(from, receiver, cpuStake, netStake))
}

// Unassign a permission currently active for the given code::action pair.
// unlinkauth [your account] [code account] [action name]
func UnlinkAuth(account, code eos.AccountName, actionName eos.ActionName) {
	api := getAPI()
	pushEOSCActions(api, system.NewUnlinkAuth(account, code, actionName))
}

// Unregister producer account temporarily.
func UnregProd(accountName eos.AccountName) {
	api := getAPI()

	pushEOSCActions(api,
		system.NewUnregProducer(accountName),
	)
}

// Unregister account as voting proxy.
func UnregProxy(accountName eos.AccountName) {
	api := getAPI()
	pushEOSCActions(api, system.NewRegProxy(accountName, false))
}

// Set or update a permission on an account. See --help for more details.
// updateauth [account] [permission_name] [parent permission or ""] [authority]
func UpdateAuth(account eos.AccountName, permissionName eos.Name, parentName, authParam string) {

	var parent eos.Name
	if parentName != "" {
		parent = toName(parentName, "parent permission")
	}

	var auth eos.Authority
	authKey, err := ecc.NewPublicKey(authParam)
	if err == nil {
		auth = eos.Authority{
			Threshold: 1,
			Keys: []eos.KeyWeight{
				{PublicKey: authKey, Weight: 1},
			},
		}
	} else {
		err := loadYAMLOrJSONFile(authParam, &auth)
		errorCheck("authority file invalid", err)
	}

	api := getAPI()

	var updateAuthActionPermission = "active"
	if parent == "" {
		updateAuthActionPermission = "owner"
	}
	pushEOSCActions(api, system.NewUpdateAuth(account, eos.PermissionName(permissionName), eos.PermissionName(parent), auth, eos.PermissionName(updateAuthActionPermission)))
}