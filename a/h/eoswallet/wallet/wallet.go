package wallet

import (
	"bytes"
	wi "gcwallet/eos-wallet-interface"
	"gcwallet/eoswallet/wallet/exchangerates"
	"gcwallet/eoswallet/config"
	"gcwallet/eoswallet/boltdb"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go"
	"fmt"
	"net/http"
	"net/http/httputil"
	"io"
	"encoding/json"
	"gcwallet/eoswallet/vault"
	"errors"
	"strings"
	"strconv"
	"gcwallet/eoswallet/eospark"
	"time"
	"github.com/shopspring/decimal"
	"math/rand"

	eosToken"github.com/eoscanada/eos-go/token"
	eosSystem "github.com/eoscanada/eos-go/system"
)

type EosWallet struct {
	pool 	*ApiPool
	vault 	*vault.Vault
	db       wi.Datastore
	eospark *eospark.API
	cAccountName eos.AccountName
	cPrivateKey  ecc.PrivateKey
	cPublicKey	 ecc.PublicKey
	cPassword 	 string
	exchangeRates wi.ExchangeRates
}

func NewEosWallet(cfg config.CoinConfig, configDir, password string, disableExchangeRates bool) (*EosWallet, error) {
	// 通过助记词生成私钥
	//seed := bip39.NewSeed(mnemonic, "")
	//mPrivKey, err:= ecc.NewDeterministicPrivateKey(bytes.NewReader(seed))
	//if err != nil {
	//	return nil, err
	//}

	//mPrivKey, err := ecc.NewPrivateKey(privateKey)
	//if err != nil {
	//	return nil, err
	//}
	//
	//mPubKey := mPrivKey.PublicKey()

	// 主网/测试网
	pool, err := NewApiPool(cfg.ClientAPIs)
	if err != nil {
		return nil, err
	}

	eosparkApiUrl, ok := cfg.Options["EosparkAPI"]
	if !ok {
		return nil, errors.New("can not find eospark api url")
	}
	eosparkApiKey, ok := cfg.Options["EosparkAPIKey"]
	if !ok {
		return nil, errors.New("can not find eospark api key")
	}

	// new eospark api
	ep := eospark.New(eosparkApiUrl.(string), eosparkApiKey.(string))
	if err != nil {
		return nil, err
	}

	// 加载一个Vault实例：从钱包文件中加载密钥、开箱（解锁）
	v, err := GetEosWalletVault(configDir, password)
	if err != nil {
		return nil, err
	}

	var (
		cAccountName eos.AccountName
		cPrivateKey  ecc.PrivateKey
		cPublicKey	 ecc.PublicKey
	)
	if len(v.KeyBag.Keys) > 0 {
		// 当KeyBag中有值时默认取第0个为当前账户
		key := v.KeyBag.Keys[0]
		account, err := cfg.DB.Accounts().Get(key.PublicKey().String())
		if err != nil {
			return nil, err
		}

		cAccountName = eos.AN(account.Name)
		cPrivateKey = *key
		cPublicKey = key.PublicKey()
	}

	var er wi.ExchangeRates
	if !disableExchangeRates {
		er = exchangerates.NewBitcoinPriceFetcher(nil)
	}

	return &EosWallet{pool: pool, vault: v, db: cfg.DB, eospark: ep, cAccountName: cAccountName, cPrivateKey: cPrivateKey, cPublicKey: cPublicKey, cPassword: password, exchangeRates: er}, nil
}

func (w *EosWallet) ImportKey(key *ecc.PrivateKey, configDir, password string)  error {
	if password != w.cPassword {
		return errors.New("password incorrect")
	}

	pubKey := key.PublicKey()

	// 校验公钥是否已存在
	if w.HasKey(pubKey) {
		return errors.New("private key already exists")
	}

	// 根据公钥获取账户名
	accounts, err := w.GetAccounts(pubKey)
	if err != nil {
		return err
	}

	// 获取账户信息
	eAccount, err := w.GetAccount(accounts.AccountNames[0])
	if err != nil {
		return err
	}

	// 获取账户权限
	authority := []string{}
	var (
		activePublicKey string
		ownerPublicKey string
	)
	for _, p := range eAccount.Permissions {
		permission := strings.Join([]string{string(eAccount.AccountName), p.PermName}, "@")
		authority = append(authority, permission)
		// 获取账户的active和owner权限
		if p.PermName == "active" {
			keys := make([]string, 0)
			for _, key := range p.RequiredAuth.Keys {
				keys = append(keys, key.PublicKey.String())
			}
			activePublicKey = strings.Join(keys, ",")
		} else if p.PermName == "owner" {
			keys := make([]string, 0)
			for _, key := range p.RequiredAuth.Keys {
				keys = append(keys, key.PublicKey.String())
			}
			ownerPublicKey = strings.Join(keys, ",")
		}
	}

	// 保存账户名和权限信息
	account := wi.Account{
		Name: string(eAccount.AccountName),
		ActivePublicKey: activePublicKey,
		OwnerPublicKey: ownerPublicKey,
		Authority: authority,
	}
	if err := w.db.Accounts().Put(account); err != nil {
		return err
	}

	// 导入私钥到vault并保存到文件
	if _, err := w.vault.AddPrivateKeyAndWriteToFile(key, configDir, password); err != nil {
		return err
	}

	// 如果是初次导入，自动切换到当前的账户
	if len(w.vault.KeyBag.Keys) == 1 {
		if err := w.ChangeAccount(pubKey); err != nil {
			return err
		}
	}

	return nil
}

// 返回所有已导入私钥的账户信息
func (w *EosWallet) ImportedKeys() ([]*wi.ImportedKey, error) {
	keybag := w.vault.KeyBag

	keys := []*wi.ImportedKey{}

	if len(keybag.Keys) != 0 {
		for _, key := range keybag.Keys {
			// 根据公钥查找账户名称
			account, err := w.db.Accounts().Get(key.PublicKey().String())
			if err != nil {
				return nil, err
			}

			activePublicKey, _ := ecc.NewPublicKey(account.ActivePublicKey)
			ownerPublicKey, _ := ecc.NewPublicKey(account.OwnerPublicKey)

			keys = append(keys, &wi.ImportedKey{
				AccountName: eos.AN(account.Name),
				ActivePublicKey: activePublicKey,
				OwnerPublicKey: ownerPublicKey,
				Authority: account.Authority,
			})
		}
	}

	return keys, nil
}

// 根据公钥切换账户为当前账户
func (w *EosWallet) ChangeAccount(pkey ecc.PublicKey) error {
	// 判断公钥是否存在
	if !w.HasKey(pkey) {
		return errors.New("private key does not exist")
	}

	// 获取账户信息
	account, err := w.db.Accounts().Get(pkey.String())
	if err != nil {
		return err
	}

	w.cAccountName = eos.AN(account.Name)

	// 根据公钥查询私钥信息
	for _, key := range w.vault.KeyBag.Keys {
		if key.PublicKey().String() == pkey.String() {
			w.cPrivateKey = *key
			w.cPublicKey = key.PublicKey()
			break
		}
	}

	return nil
}

func (w *EosWallet) Start() {
	w.pool.Start()
}

func (w *EosWallet) Params() *chaincfg.Params {
	return nil
}

func (w *EosWallet) CurrencyCode() string {
	return "EOS"
}

// 判断金额是否为尘埃
func (w *EosWallet) IsDust(amount string) bool {
	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return true
	}
	return amountInt < 1
}

func (w *EosWallet) MasterPrivateKey(password string) (*ecc.PrivateKey, error) {
	if password != w.cPassword {
		return nil, errors.New("password incorrect")
	}
	return &w.cPrivateKey, nil

}

func (w *EosWallet) MasterPublicKey() *ecc.PublicKey {
	return &w.cPublicKey
}

func (w *EosWallet) CurrentAccount() *eos.AccountName {
	return &w.cAccountName
}

func (w *EosWallet) HasKey(pkey ecc.PublicKey) bool {
	for _, key := range w.vault.KeyBag.Keys {
		if key.PublicKey().String() == pkey.String() {
			return true
		}
	}
	return false
}

func (w *EosWallet) NewKeyPair() (*ecc.PrivateKey, error) {
	key, err := ecc.NewRandomPrivateKey()
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (w *EosWallet) GetAccount(account eos.AccountName) (*eos.AccountResp, error) {
	// 返回余额的数组，格式为： ["666.6666 EOS"]
	out, err := w.pool.currentApi().GetAccount(account)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (w *EosWallet) GetAccounts(pkey ecc.PublicKey) (*wi.AccountsResp, error) {
	var out *wi.AccountsResp
	err := w.call("history", "get_key_accounts", eos.M{"public_key": pkey.String()}, &out)
	if err != nil {
		return nil, err
	}

	if len(out.AccountNames) == 0 {
		return nil, fmt.Errorf("account names is nil")
	}
	return out, err
}

func (w *EosWallet) GetBalance(account eos.AccountName) int64 {
	// 返回余额的数组，格式为： ["666.6666 EOS"]
	out, err := w.pool.currentApi().GetCurrencyBalance(account, w.CurrencyCode(), "eosio.token")
	if err != nil {
		return 0
	}
	if len(out) == 0 {
		return 0
	}

	return out[0].Amount
}

func (w *EosWallet) GetTransactions(transactionType, sort, page, size int) ([]*wi.Txn, int32, error) {
	txns := make([]*wi.Txn, 0)

	// 官方不提供获取所有交易接口，所以调用eospark接口获取
	resp, err := w.eospark.GetAccountRelatedTrxInfo(string(w.cAccountName), transactionType, sort, page, size)
	if err != nil {
		return nil, 0, err
	}

	// 获取当前区块高度
	chainTip, _ := w.ChainTip()

	for _, tx := range resp.TraceList {
		txn := &wi.Txn{
			Txid: tx.TrxId,
			Height: int32(tx.BlockNum),
			Timestamp: tx.Timestamp.Time,
			Symbol: tx.Symbol,
			Status: tx.Status,
			Sender: string(tx.Sender),
			Receiver: string(tx.Receiver),
			Memo: tx.Memo,
		}

		// 计算交易确认数
		txn.Confirmations = int32(chainTip) - txn.Height

		amount, err := strconv.ParseFloat(tx.Quantity, 64)
		if err != nil {
			return nil, 0, err
		}

		// 将金额乘以10000取整
		var value string = decimal.NewFromFloat(amount).Mul(decimal.NewFromFloat(10000)).String()
		// 根据转账接收者判断该交易是支出还是收入
		if string(tx.Receiver) != string(w.cAccountName) {
			value = fmt.Sprintf("-%s", value)
		}
		txn.Value = value

		txns = append(txns, txn)
	}

	return txns, resp.TraceCount, nil
}

func (w *EosWallet) GetTransaction(txid chainhash.Hash) (*wi.Txn, error) {
	// 获取本次交易信息, 有些api没开插件，所以用eospark来获取
	resp, err := w.eospark.GetTransactionDetailInfo(txid.String())
	if err != nil {
		return nil, err
	}

	// 获取当前区块高度
	chainTip, _ := w.ChainTip()

	txn := &wi.Txn{
		Height: int32(resp.BlockNum),
		Timestamp: resp.Timestamp.Time,
		Confirmations: int32(chainTip) - int32(resp.BlockNum),
	}

	switch resp.EosparkTrxType {
	case "ordinary":
		txn.Txid = resp.Trx.ID.String()
		txn.Status = resp.Status.String()

		for _, tx := range resp.Trx.Transaction.Actions {
			if tx.Name == "transfer" {
				data := tx.Data.(map[string]interface{})
				// 如果是transfer操作 交易金额保存在data的quantity字段中，格式为 "666.6666 EOS"， 所以需要分割取出金额
				quantity, ok := data["quantity"]
				if ok {
					ql := strings.Split(quantity.(string), " ")
					if len(ql) == 2 {
						amount, err := strconv.ParseFloat(ql[0], 64)
						if err != nil {
							return nil, err
						}
						txn.Sender, _ = data["from"].(string)
						txn.Receiver, _ = data["to"].(string)
						txn.Memo, _ = data["memo"].(string)

						// 将金额乘以10000取整
						var value string = decimal.NewFromFloat(amount).Mul(decimal.NewFromFloat(10000)).String()
						// 根据转账接收者判断该交易是支出还是收入
						if txn.Receiver != string(w.cAccountName) {
							value = fmt.Sprintf("-%s", value)
						}
						txn.Value = value
						txn.Symbol = ql[1]
					}
				}
			}
		}
		return txn, nil
	case "inline":
		txn.Txid = resp.ID.String()
		txn.Status = resp.Trx.Receipt.Status.String()

		for _, trace := range resp.Traces {
			if trace.Action.Name == "transfer" {
				data := trace.Action.Data.(map[string]interface{})
				// 如果是transfer操作 交易金额保存在data的quantity字段中，格式为 "666.6666 EOS"， 所以需要分割取出金额
				quantity, ok := data["quantity"]
				if ok {
					ql := strings.Split(quantity.(string), " ")
					if len(ql) == 2 {
						amount, err := strconv.ParseFloat(ql[0], 64)
						if err != nil {
							return nil, err
						}
						txn.Sender, _ = data["from"].(string)
						txn.Receiver, _ = data["to"].(string)
						txn.Memo, _ = data["memo"].(string)

						// 将金额乘以10000取整
						var value string = decimal.NewFromFloat(amount).Mul(decimal.NewFromFloat(10000)).String()
						// 根据转账接收者判断该交易是支出还是收入
						if txn.Receiver != string(w.cAccountName) {
							value = fmt.Sprintf("-%s", value)
						}
						txn.Value = value
						txn.Symbol = ql[1]
						break
					}
				}
			}
		}
		return txn, nil
	}

	return nil, errors.New("not found")
}

func (w *EosWallet) ChainTip() (uint32, *chainhash.Hash) {
	info, err := w.pool.currentApi().GetInfo()
	h, _ := chainhash.NewHashFromStr("")
	if err != nil {
		return 0, h
	}
	
	h, _ = chainhash.NewHashFromStr(info.HeadBlockID.String())
	return info.LastIrreversibleBlockNum, h
}

func (w *EosWallet) GetAccountResourceInfo() (*eospark.GetAccountResourceInfoResp, error) {
	resp, err := w.eospark.GetAccountResourceInfo(string(w.cAccountName))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (w *EosWallet) GetTokenList() ([]*eospark.Symbol, error) {
	resp, err := w.eospark.GetTokenList(string(w.cAccountName), "")
	if err != nil {
		return nil, err
	}

	return resp.SymbolList, nil
}

// 发送一个EOS交易
func (w *EosWallet) Spend(to eos.AccountName, memo, amount string) (*chainhash.Hash, error) {
	api := w.pool.currentApi()
	keyBag := w.vault.KeyBag

	// 使用KeyBag作为签名提供者
	api.SetSigner(keyBag)

	// 解析交易金额
	amountInt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		return nil, err
	}

	quantity := eos.NewEOSAsset(amountInt)

	// 创建交易操作
	action := eosToken.NewTransfer(w.cAccountName, to, quantity, memo)

	opts := &eos.TxOptions{}

	if err := opts.FillFromChain(api); err != nil {
		return nil, err
	}

	// 新交易
	tx := eos.NewTransaction([]*eos.Action{action}, opts)
	tx.SetExpiration(60 * time.Second)

	// 获取签名Key
	if len(keyBag.Keys) > 0 {
		var signKeys []ecc.PublicKey
		for _, key := range w.vault.KeyBag.Keys {
			signKeys = append(signKeys, key.PublicKey())
		}
		// 设置定制获取要求的Keys
		api.SetCustomGetRequiredKeys(func(tx *eos.Transaction) ([]ecc.PublicKey, error) {
			return signKeys, nil
		})
	}

	// 将交易签名并发送操作
	out, err := api.SignPushTransaction(tx, opts.ChainID, opts.Compress)
	if err != nil {
		return nil, err
	}

	hash, err := chainhash.NewHashFromStr(out.TransactionID)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (w *EosWallet) NewAccountName() (eos.AccountName, error) {
	accountRule := []byte("abcdefghijklmnopqrstuvwxyz12345")

	bytesLen := len(accountRule)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 10; i++ {
		result = append(result, accountRule[r.Intn(bytesLen)])
	}

	acccountName := eos.AN("gc" + string(result))

	// 如果随机生成的账户未使用就返回，否则继续生成
	if _, err := w.GetAccount(acccountName); err != nil {
		return acccountName, nil
	} else {
		return w.NewAccountName()
	}
}

var galaxyCipherText = `SSOkY0kLUVH1kvGehCZy8MWi/w0btu224NhsK+boYPQkZQOcDE8B1ub5jDZ0jHE6Dy2ukAGNQ9NioHmfwyjzLkK0uIpVJgWzX2ZI0gE5uRmH6f1pYHO7dHy6960q+iXUdWgtYeN9WXPywmbnag`

// GalaxyCoin 创建一个新账户
func (w *EosWallet) NewAccountByGC(publicKey ecc.PublicKey, name eos.AccountName) (*chainhash.Hash, error) {
	api := w.pool.currentApi()

	// 如果账户名为空，则获取一个新的随机EOS账户名
	if name == "" {
		newName, err := w.NewAccountName()
		if err != nil {
			return nil, err
		}
		name = newName
	}

	p := vault.NewPassphraseBoxer("U2FsdGVkX1+ncdCbLnf9F5P3WrXp6eULQAFAK/U+VmZ6fnapDnBcioCwDGZfao/+")

	//payload, err := json.Marshal("")
	//if err != nil {
	//	return nil, err
	//}
	//cipherText, err := p.Seal(payload)
	//if err != nil {
	//	return nil, err
	//}

	// 解析密文获取私钥
	cipherByte, err := p.Open(galaxyCipherText)
	if err != nil {
		return nil, err
	}
	var gKey string
	if err := json.Unmarshal(cipherByte, &gKey); err != nil {
		return nil, err
	}

	v := vault.NewVault()
	gPrivateKey, err := ecc.NewPrivateKey(gKey)
	if err != nil {
		return nil, err
	}
	v.AddPrivateKey(gPrivateKey)

	// 使用KeyBag作为签名提供者
	api.SetSigner(v.KeyBag)

	creator := eos.AN("huangjingshu")

	var actions []*eos.Action

	// 创建账户操作
	actions = append(actions, eosSystem.NewNewAccount(creator, name, publicKey))
	// 分配内存
	actions = append(actions, eosSystem.NewBuyRAM(creator, name, 3000))
	// 分配CPU和NET
	actions = append(actions, eosSystem.NewDelegateBW(creator, name, eos.NewEOSAsset(1), eos.NewEOSAsset(1), false))

	opts := &eos.TxOptions{}

	if err := opts.FillFromChain(api); err != nil {
		return nil, err
	}

	// 新交易
	tx := eos.NewTransaction(actions, opts)
	tx.SetExpiration(180 * time.Second)

	// 获取签名Key
	if len(v.KeyBag.Keys) > 0 {
		var signKeys []ecc.PublicKey
		for _, key := range v.KeyBag.Keys {
			signKeys = append(signKeys, key.PublicKey())
		}
		// 设置定制获取要求的Keys
		api.SetCustomGetRequiredKeys(func(tx *eos.Transaction) ([]ecc.PublicKey, error) {
			return signKeys, nil
		})
	}

	// 将交易签名并发送操作
	out, err := api.SignPushTransaction(tx, opts.ChainID, opts.Compress)
	if err != nil {
		return nil, err
	}

	hash, err := chainhash.NewHashFromStr(out.TransactionID)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func (w *EosWallet) ExchangeRates() wi.ExchangeRates {
	return w.exchangeRates
}

func (w *EosWallet) Close() {
	if w.db != nil{
		db := w.db.(*boltdb.BoltDatastore)
		db.DB.Close()
	}
}

// 由于有些接口eos-go并未提供，所以提供该方法自行请求
func (w *EosWallet) call(baseAPI string, endpoint string, body interface{}, out interface{}) error {
	api := w.pool.currentApi()

	v, err := json.Marshal(body)
	if err != nil {
		return err
	}

	jsonBody := bytes.NewReader(v)
	if err != nil {
		return err
	}

	targetURL := fmt.Sprintf("%s/v1/%s/%s", api.BaseURL, baseAPI, endpoint)
	req, err := http.NewRequest("POST", targetURL, jsonBody)
	if err != nil {
		return fmt.Errorf("NewRequest: %s", err)
	}

	for k, v := range api.Header {
		if req.Header == nil {
			req.Header = http.Header{}
		}
		req.Header[k] = append(req.Header[k], v...)
	}

	if api.Debug {
		// Useful when debugging API calls
		requestDump, err := httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(string(requestDump))
		fmt.Println("")
	}

	resp, err := api.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %s", req.URL.String(), err)
	}
	defer resp.Body.Close()

	var cnt bytes.Buffer
	_, err = io.Copy(&cnt, resp.Body)
	if err != nil {
		return fmt.Errorf("Copy: %s", err)
	}

	if resp.StatusCode == 404 {
		var apiErr eos.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return eos.ErrNotFound
		}
		return apiErr
	}
	if resp.StatusCode > 299 {
		var apiErr eos.APIError
		if err := json.Unmarshal(cnt.Bytes(), &apiErr); err != nil {
			return fmt.Errorf("%s: status code=%d, body=%s", req.URL.String(), resp.StatusCode, cnt.String())
		}
		return apiErr
	}

	if api.Debug {
		fmt.Println("RESPONSE:")
		responseDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("-------------------------------")
		fmt.Println(cnt.String())
		fmt.Println("-------------------------------")
		fmt.Printf("%q\n", responseDump)
		fmt.Println("")
	}

	if err := json.Unmarshal(cnt.Bytes(), &out); err != nil {
		return fmt.Errorf("Unmarshal: %s", err)
	}

	return nil
}
