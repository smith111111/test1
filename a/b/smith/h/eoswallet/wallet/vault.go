package wallet

import (
	"encoding/json"
	"fmt"
	eosvault "gcwallet/eoswallet/vault"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"path"
)

// Add private keys to an existing vault taking input from the shell
func VaultAdd() {
	// 钱包文件名
	walletFile := viper.GetString("global-vault-file")

	fmt.Println("Loading existing vault from file:", walletFile)
	// 从提供的eos钱包文件名返回一个新的Vault实例。
	vault, err := eosvault.NewVaultFromWalletFile(walletFile)
	errorCheck("loading vault from file", err)

	// 返回一个指定类型的秘密装箱器
	boxer, err := eosvault.SecretBoxerForType(vault.SecretBoxWrap, "password")
	errorCheck("missing parameters", err)

	// 开箱：解密Vault中的密码箱密文SecretBoxCiphertext，解密的结果反序列化到KeyBag中去。
	err = vault.Open(boxer)
	errorCheck("opening vault", err)

	// 打印Vault的密钥袋KeyBag中的每个私钥的公钥。
	vault.PrintPublicKeys()

	// 粘贴输入私钥串
	privateKeys, err := capturePrivateKeys()
	errorCheck("entering private keys", err)

	var newKeys []ecc.PublicKey
	for _, privateKey := range privateKeys {
		vault.AddPrivateKey(privateKey)
		newKeys = append(newKeys, privateKey.PublicKey())
	}

	// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
	err = vault.Seal(boxer)
	errorCheck("sealing vault", err)

	// 将Vault写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
	err = vault.WriteToFile(walletFile)
	errorCheck("writing vault file", err)

	vaultWrittenReport(walletFile, newKeys, len(vault.KeyBag.Keys))
}

func capturePrivateKeys() (out []*ecc.PrivateKey, err error) {
	fmt.Println("")
	fmt.Println("PLEASE READ:")
	fmt.Println("We are now going to ask you to paste your private keys, one at a time.")
	fmt.Println("They will not be shown on screen.")
	fmt.Println("Please verify that the public keys printed on screen correspond to what you have noted")
	fmt.Println("")

	first := true
	for {
		privKey, err := capturePrivateKey(first)
		if err != nil {
			return out, fmt.Errorf("capture privkeys: %s", err)
		}
		first = false

		if privKey == nil {
			return out, nil
		}
		out = append(out, privKey)
	}
}

func capturePrivateKey(isFirst bool) (privateKey *ecc.PrivateKey, err error) {
	prompt := "Paste your first private key: "
	if !isFirst {
		prompt = "Paste your next private key or hit ENTER if you are done: "
	}

	enteredKey, err := GetPassword(prompt)
	if err != nil {
		return nil, fmt.Errorf("get private key: %s", err)
	}

	if enteredKey == "" {
		return nil, nil
	}

	key, err := ecc.NewPrivateKey(enteredKey)
	if err != nil {
		return nil, fmt.Errorf("import private key: %s", err)
	}

	fmt.Printf("- Scanned private key corresponding to %s\n", key.PublicKey().String())

	return key, nil
}

func vaultWrittenReport(walletFile string, newKeys []ecc.PublicKey, totalKeys int) {
	fmt.Println("")
	fmt.Printf("Wallet file %q written to disk.\n", walletFile)
	fmt.Println("Here are the keys that were ADDED during this operation (use `list` to see them all):")
	for _, pub := range newKeys {
		fmt.Printf("- %s\n", pub.String())
	}

	fmt.Printf("Total keys stored: %d\n", totalKeys)
}

// Create a new encrypted EOS keys vault
/*
	A vault contains encrypted private keys, and with 'eosc', can be used to
	securely sign transactions.

	You can create a passphrase protected vault with:

		eosc vault create --keys=2

	This uses the default --vault-type=passphrase

	You can create a Google Cloud Platform KMS-wrapped vault with:

		eosc vault create --keys=2 --vault-type=kms-gcp --kms-gcp-keypath projects/.../locations/.../keyRings/.../cryptoKeys/name

	You can then use this vault for the different eosc operations.`
*/
func VaultCreate() {
	// 钱包文件名
	walletFile := viper.GetString("global-vault-file")

	// 如果钱包文件已经存在了，则退出
	if _, err := os.Stat(walletFile); err == nil {
		fmt.Printf("Wallet file %q already exists, rename it before running `eosc vault create`.\n", walletFile)
		os.Exit(1)
	}

	// 钱包密封器类型
	var wrapType = viper.GetString("vault-create-cmd-vault-type")
	var boxer eosvault.SecretBoxer

	// 返回一个空vault，未存储并且不含密钥。
	vault := eosvault.NewVault()
	vault.Comment = viper.GetString("vault-create-cmd-comment")

	var newKeys []ecc.PublicKey

	// 如果要导入私钥，则
	doImport := viper.GetBool("vault-create-cmd-import")
	if doImport {
		// 提示粘贴导入私钥
		privateKeys, err := capturePrivateKeys()
		errorCheck("entering private key", err)

		for _, privateKey := range privateKeys {
			vault.AddPrivateKey(privateKey)
			newKeys = append(newKeys, privateKey.PublicKey())
		}

		fmt.Printf("Imported %d keys.\n", len(newKeys))

	} else { // 否则，要创建n个密钥对，并导入新私钥。
		numKeys := viper.GetInt("vault-create-cmd-keys")

		if numKeys == 0 {
			errorCheck("specify either --keys or --import", fmt.Errorf("create a vault with 0 keys?"))
		}

		for i := 0; i < numKeys; i++ {
			// 创建一个新的EOS密钥对，保存私钥在本地钱包中，并返回公钥。它不存储这个钱包，你最好是马上做保存。
			pubKey, err := vault.NewKeyPair()
			errorCheck("creating new keypair", err)

			newKeys = append(newKeys, pubKey)
		}
		fmt.Printf("Created %d keys. They will be shown when encrypted and written to disk successfully.\n", len(newKeys))
	}

	switch wrapType {
		case "passphrase":
			fmt.Println("")
			fmt.Println("You will be asked to provide a passphrase to secure your newly created vault.")
			fmt.Println("Make sure you make it long and strong.")
			fmt.Println("")
			if envVal := os.Getenv("EOSC_GLOBAL_INSECURE_VAULT_PASSPHRASE"); envVal != "" {
				// 新建一个密码装箱器实例
				boxer = eosvault.NewPassphraseBoxer(envVal)
			} else {
				// 提示输入一个密码
				password, err := GetEncryptPassphrase()
				errorCheck("password input", err)

				// 新建一个密码装箱器实例
				boxer = eosvault.NewPassphraseBoxer(password)
			}
		default:
			fmt.Printf(`Invalid vault type: %q, please use one of: "passphrase", "kms-gcp"\n`, wrapType)
			os.Exit(1)
	}

	// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
	errorCheck("sealing vault", vault.Seal(boxer))
	// 将Vault写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
	errorCheck("writing wallet file", vault.WriteToFile(walletFile))

	vaultWrittenReport(walletFile, newKeys, len(vault.KeyBag.Keys))
}

func GetEosWalletVault(configDir, password string) (*eosvault.Vault, error) {
	// 配置文件包文件名
	walletFile := path.Join(configDir, ".gcwallet", "gcwallet-eos")
	// 如果钱包文件已经存在了，则退出
	if _, err := os.Stat(walletFile); err == nil {
		// 从提供的eos钱包文件名返回一个新的Vault实例。
		v, err := eosvault.NewVaultFromWalletFile(walletFile)
		if err != nil {
			return nil, fmt.Errorf("new vault from wallet file error: %s", err.Error())
		}

		// 获取秘密装箱器
		boxer, err := eosvault.SecretBoxerForType(v.SecretBoxWrap, password)
		if err != nil {
			return nil, fmt.Errorf("secret boxer error: %s", err.Error())
		}

		// 开箱：（填充KeyBag属性）
		if err := v.Open(boxer); err != nil {
			return nil, fmt.Errorf("open vault error: %s, password is %s, SecretBoxCiphertext is %s", err.Error(), password, v.SecretBoxCiphertext)
		}

		return v, nil
	}

	// 返回一个空vault，未存储并且不含密钥。
	v := eosvault.NewVault()

	var boxer eosvault.SecretBoxer
	boxer = eosvault.NewPassphraseBoxer(password)

	// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
	if err := v.Seal(boxer); err != nil {
		return nil, err
	}
	// 将Vault写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
	if err := v.WriteToFile(walletFile); err != nil {
		return nil, err
	}

	return v, nil
}

// Export private keys (and corresponding public keys) inside an eosc vault.
func VaultExport() {
	// 加载一个Vault实例：从钱包文件中加载密钥、开箱（解锁）。
	vault := mustGetWallet()

	vault.PrintPrivateKeys()
}

// List public keys inside an eosc vault.
func VaultList() {
	vault := mustGetWallet()

	vault.PrintPublicKeys()
}

// Serve will start listening on a local port, offering a
// keosd-compatible interface, ready to sign transactions.
//
// It is to be used with tools such as 'cleos' or 'eos-vote' that need
// transactions signed before submitting them to an EOS network.
func VaultServe(accountName eos.AccountName) {
	vault := mustGetWallet()

	vault.PrintPublicKeys()

	listen(vault)
}

func listen(v *eosvault.Vault) {
	// 获取钱包中所有的公钥
	http.HandleFunc("/v1/wallet/get_public_keys", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("Service /v1/wallet/get_public_keys")

		var out []string
		for _, key := range v.KeyBag.Keys {
			out = append(out, key.PublicKey().String())
		}
		json.NewEncoder(w).Encode(out)
	})

	// 进来的签名请求
	http.HandleFunc("/v1/wallet/sign_transaction", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Incoming signature request")

		var inputs []json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&inputs); err != nil {
			fmt.Println("sign_transaction: error:", err)
			http.Error(w, "couldn't decode input", 500)
			return
		}

		var tx *eos.SignedTransaction
		var requiredKeys []ecc.PublicKey
		var chainID eos.HexBytes

		if len(inputs) != 3 {
			http.Error(w, "invalid length of message, should be 3 parameters", 500)
			return
		}

		// 要签名的交易
		err := json.Unmarshal(inputs[0], &tx)
		if err != nil {
			http.Error(w, "decoding transaction", 500)
			return
		}

		// 要求的公钥串数组
		err = json.Unmarshal(inputs[1], &requiredKeys)
		if err != nil {
			http.Error(w, "decoding required keys", 500)
			return
		}

		// 链ID
		err = json.Unmarshal(inputs[2], &chainID)
		if err != nil {
			http.Error(w, "decoding chain id", 500)
			return
		}

		fmt.Println("")

		// 如果不是自动接受，则
		if !viper.GetBool("vault-serve-cmd-auto-accept") {
			// 提示输入一个随机的验证码
			res, err := GetConfirmation(`- Enter the code "%d" to allow signature: `)
			if err != nil {
				fmt.Println("sign_transaction: error reading confirmation from command line:", err)
				http.Error(w, "error reading confirmation from command line", 500)
				return
			}

			if !res {
				fmt.Println("sign_transaction: security code invalid, not signing request")
				http.Error(w, "security code invalid, not signing request", 401)
				return
			}
		} else { // 否则，是自动签名，无需确认。
			fmt.Println("- Auto-signing request")
		}

		// 签名一个交易
		signed, err := v.KeyBag.Sign(tx, chainID, requiredKeys...)
		for _, action := range signed.Transaction.Actions {
			// 不发送到EOS网络上去，即，只是签名而已。
			action.SetToServer(false)
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("error signing: %s", err), 500)
			return
		}

		// 转换成JSON串字节片段
		cnt, err := json.Marshal(signed)
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't marshal output: %s", err), 500)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(cnt)))
		w.WriteHeader(201)
		// 将签名的交易发送给客户端
		_, err = w.Write(cnt)
		if err != nil {
			log.Println("Error writing to socket:", err)
		}
	})

	port := viper.GetInt("vault-serve-cmd-port")
	fmt.Printf("Listening for wallet operations on 127.0.0.1:%d\n", port)
	// 启动http监听
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil); err != nil {
		fmt.Printf("Failed listening on port %d: %s\n", port, err)
	}
}
