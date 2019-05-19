package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	eosvault "gcwallet/eoswallet/vault"
	"github.com/bronze1man/go-yaml2json"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/sudo"
	"github.com/spf13/viper"
	"github.com/tidwall/sjson"
)

var reValidAccount = regexp.MustCompile(`[a-z12345]+`)

// ToAccountName converts a eos valid name string (in) into an eos-go
// AccountName struct
func ToAccountName(in string) (out eos.AccountName, err error) {
	if !reValidAccount.MatchString(in) {
		err = fmt.Errorf("invalid characters in %q, allowed: 'a' through 'z', and '1', '2', '3', '4', '5'", in)
		return
	}

	if len(in) > 12 {
		err = fmt.Errorf("%q too long, 12 characters allowed maximum", in)
		return
	}

	if len(in) == 0 {
		err = fmt.Errorf("empty")
		return
	}

	return eos.AccountName(in), nil
}

// ToName converts a valid eos name string (in) into an eos-go
// Name struct
func ToName(in string) (out eos.Name, err error) {
	name, err := ToAccountName(in)
	if err != nil {
		return
	}
	return eos.Name(name), nil
}

// 加载一个Vault实例：从钱包文件中加载密钥、开箱（解锁）。
func mustGetWallet() *eosvault.Vault {
	// 初始化Vault实例
	vault, err := setupWallet()
	// 如果出错，则退出1
	errorCheck("wallet setup", err)
	// 返回Vault实例
	return vault
}

func setupWallet() (*eosvault.Vault, error) {
	walletFile := viper.GetString("vault.wallet_file")
	vaultPassword := viper.GetString("vault.password")

	if _, err := os.Stat(walletFile); err != nil {
		// 丢失钱包文件，则
		return nil, fmt.Errorf("wallet file %q missing: %s", walletFile, err)
	}

	// 从提供的eos钱包文件名返回一个新的Vault实例。
	vault, err := eosvault.NewVaultFromWalletFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("loading vault: %s", err)
	}

	// 获取秘密装箱器
	boxer, err := eosvault.SecretBoxerForType(vault.SecretBoxWrap, vaultPassword)
	if err != nil {
		return nil, fmt.Errorf("secret boxer: %s", err)
	}

	// 开箱：（填充KeyBag属性）
	if err := vault.Open(boxer); err != nil {
		return nil, err
	}

	return vault, nil
}

func attachWallet(api *eos.API) {
	walletURLs := []string{};//viper.GetStringSlice("global-wallet-url")
	if len(walletURLs) == 0 {
		vault, err := setupWallet()
		errorCheck("setting up wallet", err)

		// 使用KeyBag作为签名提供者
		api.SetSigner(vault.KeyBag)
	} else {
		if len(walletURLs) == 1 {
			// If a `walletURLs` has a Username in the path, use instead of `default`.
			// 使用keosd提供的嵌套钱包：默认default钱包。
			api.SetSigner(eos.NewWalletSigner(eos.New(walletURLs[0]), "default"))
		} else {
			fmt.Println("Multi-signer not yet implemented.  Please choose only one `--wallet-url`")
			os.Exit(1)
		}
	}
}

func getAPI() *eos.API {
	httpHeaders := []string{};//viper.GetStringSlice("global-http-header")
	// 使用测试网
	api := eos.New("http://kylin.meet.one:8888") //viper.GetString("global-api-url"))
	for _, header := range httpHeaders {
		headerArray := strings.SplitN(header, ": ", 2)
		if len(headerArray) != 2 || strings.Contains(headerArray[0], " ") {
			errorCheck("validating http headers", fmt.Errorf("invalid HTTP Header format"))
		}
		api.Header.Add(headerArray[0], headerArray[1])
	}
	return api
}

// 如果有错误，则退出进程
func errorCheck(prefix string, err error) {
	if err != nil {
		fmt.Printf("ERROR: %s: %s\n", prefix, err)
		os.Exit(1)
	}
}

// 解析一个权限串（`account@active`、`otheraccount@owner`）
func permissionToPermissionLevel(in string) (out eos.PermissionLevel, err error) {
	// 解析`account@active`、`otheraccount@owner`，并建立一个PermissionLevel结构。
	return eos.NewPermissionLevel(in)
}

// 解析多个权限串
func permissionsToPermissionLevels(in []string) (out []eos.PermissionLevel, err error) {
	// 遍历所有参数
	for _, singleArg := range in {

		// 如果指定了"account@active,account2"，也处理它。
		for _, val := range strings.Split(singleArg, ",") {
			level, err := permissionToPermissionLevel(strings.TrimSpace(val))
			if err != nil {
				return out, err
			}

			out = append(out, level)
		}
	}

	return
}

// 推送EOS动作（可多个）
func pushEOSCActions(api *eos.API, actions ...*eos.Action) {
	pushEOSCActionsAndContextFreeActions(api, nil, actions)
}

// 推送EOS动作和上下文免费动作
func pushEOSCActionsAndContextFreeActions(api *eos.API, contextFreeActions []*eos.Action, actions []*eos.Action) {
	for _, act := range contextFreeActions {
		// 免费动作，无需权限
		act.Authorization = nil
	}

	permissions := []string{"zhengjun1111@active","zhengjun1112@action"}// viper.GetStringSlice("global-permission")
	if len(permissions) != 0 {
		levels, err := permissionsToPermissionLevels(permissions)
		errorCheck("specified --permission(s) invalid", err)

		for _, act := range actions {
			// 设置动作权限
			act.Authorization = levels
		}
	}

	opts := &eos.TxOptions{}

	if chainID := viper.GetString("global-offline-chain-id"); chainID != "" {
		opts.ChainID = toSHA256Bytes(chainID, "--offline-chain-id")
	}

	if headBlockID := viper.GetString("global-offline-head-block"); headBlockID != "" {
		opts.HeadBlockID = toSHA256Bytes(headBlockID, "--offline-head-block")
	}

	if delaySec := viper.GetInt("global-delay-sec"); delaySec != 0 {
		opts.DelaySecs = uint32(delaySec)
	}

	if err := opts.FillFromChain(api); err != nil {
		fmt.Println("Error fetching tapos + chain_id from the chain (specify --offline flags for offline operations):", err)
		os.Exit(1)
	}

	// 新交易
	tx := eos.NewTransaction(actions, opts)
	// 如果有上下文免费动作的话，则
	if len(contextFreeActions) > 0 {
		// 设置上下文免费动作
		tx.ContextFreeActions = contextFreeActions
	}

	// 可选地进行sudo封装
	tx = optionallySudoWrap(tx, opts)

	// 有效期
	tx.SetExpiration(time.Duration(viper.GetInt("global-expiration")) * time.Second)

	// 可选地签名这个交易。
	signedTx, packedTx := optionallySignTransaction(tx, opts.ChainID, api)

	// 可选地推送这个交易。
	optionallyPushTransaction(signedTx, packedTx, opts.ChainID, api)
}

func optionallySudoWrap(tx *eos.Transaction, opts *eos.TxOptions) *eos.Transaction {
	if viper.GetBool("global-sudo-wrap") {
		return eos.NewTransaction([]*eos.Action{sudo.NewExec(eos.AccountName("eosio"), *tx)}, opts)
	}
	return tx
}

func optionallySignTransaction(tx *eos.Transaction, chainID eos.SHA256Bytes, api *eos.API) (signedTx *eos.SignedTransaction, packedTx *eos.PackedTransaction) {
	// 如果不跳过签名的话，则
	if !viper.GetBool("global-skip-sign") {
		// 签名keys串[]string{}
		textSignKeys := viper.GetStringSlice("global-offline-sign-key")
		if len(textSignKeys) > 0 {
			var signKeys []ecc.PublicKey
			for _, key := range textSignKeys {
				pubKey, err := ecc.NewPublicKey(key)
				errorCheck(fmt.Sprintf("parsing public key %q", key), err)

				signKeys = append(signKeys, pubKey)
			}
			// 设置定制获取要求的Keys
			api.SetCustomGetRequiredKeys(func(tx *eos.Transaction) ([]ecc.PublicKey, error) {
				return signKeys, nil
			})
		}

		// 设置签名提供者
		attachWallet(api)

		var err error
		// 签名交易
		signedTx, packedTx, err = api.SignTransaction(tx, chainID, eos.CompressionNone)
		errorCheck("signing transaction", err)
	} else {
		signedTx = eos.NewSignedTransaction(tx)
	}
	return signedTx, packedTx
}

func optionallyPushTransaction(signedTx *eos.SignedTransaction, packedTx *eos.PackedTransaction, chainID eos.SHA256Bytes, api *eos.API) {
	writeTrx := viper.GetString("global-write-transaction")

	// 如果要写入到一个交易文件中去的话，
	if writeTrx != "" {
		// 签名的交易序列化成JSON字节串
		cnt, err := json.MarshalIndent(signedTx, "", "  ")
		errorCheck("marshalling json", err)

		// 设置chain_id属性
		annotatedCnt, err := sjson.Set(string(cnt), "chain_id", hex.EncodeToString(chainID))
		errorCheck("adding chain_id", err)

		// 写入到交易文件中去
		err = ioutil.WriteFile(writeTrx, []byte(annotatedCnt), 0644)
		errorCheck("writing output transaction", err)

		fmt.Printf("Transaction written to %q\n", writeTrx)
	} else {
		if packedTx == nil {
			fmt.Println("A signed transaction is required if you want to broadcast it. Remove --skip-sign (or add --write-transaction ?)")
			os.Exit(1)
		}

		// TODO: print the traces
		// 推送这个交易。
		pushTransaction(api, packedTx, chainID)
	}
}

func pushTransaction(api *eos.API, packedTx *eos.PackedTransaction, chainID eos.SHA256Bytes) {
	// 推送交易
	resp, err := api.PushTransaction(packedTx)
	errorCheck("pushing transaction", err)

	//fmt.Println("Transaction submitted to the network. Confirm at https://eosq.app/tx/" + resp.TransactionID)
	trxURL := transactionURL(chainID, resp.TransactionID)
	fmt.Printf("\nTransaction submitted to the network.\n  %s\n", trxURL)
	if resp.BlockID != "" {
		blockURL := blockURL(chainID, resp.BlockID)
		fmt.Printf("Server says transaction was included in block %d:\n  %s\n", resp.BlockNum, blockURL)
	}
}

func transactionURL(chainID eos.SHA256Bytes, trxID string) string {
	hexChain := hex.EncodeToString(chainID)
	switch hexChain {
	case "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906":
		return fmt.Sprintf("https://eosq.app/tx/%s", trxID)
	case "5fff1dae8dc8e2fc4d5b23b2c7665c97f9e9d8edf2b6485a86ba311c25639191":
		return fmt.Sprintf("https://kylin.eosq.app/tx/%s", trxID)
	}
	return trxID
}

func blockURL(chainID eos.SHA256Bytes, blockID string) string {
	hexChain := hex.EncodeToString(chainID)
	switch hexChain {
	case "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906":
		return fmt.Sprintf("https://eosq.app/block/%s", blockID)
	case "5fff1dae8dc8e2fc4d5b23b2c7665c97f9e9d8edf2b6485a86ba311c25639191":
		return fmt.Sprintf("https://kylin.eosq.app/block/%s", blockID)
	}
	return blockID
}

func yamlUnmarshal(cnt []byte, v interface{}) error {
	jsonCnt, err := yaml2json.Convert(cnt)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonCnt, v)
}

func loadYAMLOrJSONFile(filename string, v interface{}) error {
	cnt, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return json.Unmarshal(cnt, v)
	}
	return yamlUnmarshal(cnt, v)
}

func toAccount(in, field string) eos.AccountName {
	acct, err := ToAccountName(in)
	if err != nil {
		errorCheck(fmt.Sprintf("invalid account format for %q", field), err)
	}

	return acct
}

func toName(in, field string) eos.Name {
	name, err := ToName(in)
	if err != nil {
		errorCheck(fmt.Sprintf("invalid name format for %q", field), err)
	}

	return name
}

func toPermissionLevel(in, field string) eos.PermissionLevel {
	perm, err := permissionToPermissionLevel(in)
	if err != nil {
		errorCheck(fmt.Sprintf("invalid permission level for %q", field), err)
	}
	return perm
}

func toActionName(in, field string) eos.ActionName {
	return eos.ActionName(toName(in, field))
}

func toSHA256Bytes(in, field string) eos.SHA256Bytes {
	if len(in) != 64 {
		errorCheck(fmt.Sprintf("%q invalid", field), errors.New("should be 64 hexadecimal characters"))
	}

	bytes, err := hex.DecodeString(in)
	errorCheck(fmt.Sprintf("invalid hex in %q", field), err)

	return bytes
}

// 是一个空ABI
func isStubABI(abi eos.ABI) bool {
	return abi.Version == "" &&
		abi.Actions == nil &&
		abi.ErrorMessages == nil &&
		abi.Extensions == nil &&
		abi.RicardianClauses == nil &&
		abi.Structs == nil && abi.Tables == nil &&
		abi.Types == nil
}
