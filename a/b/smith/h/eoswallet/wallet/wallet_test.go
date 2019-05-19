package wallet

import (
	eosconfig "gcwallet/eoswallet/config"
	wi "gcwallet/eos-wallet-interface"
	eosdb "gcwallet/eoswallet/boltdb"
	"os"
	"testing"
	"path"
	"github.com/eoscanada/eos-go"
	"github.com/spf13/viper"
	"galaxyotc/common/utils"
	"strconv"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

const (
	// test11111112
	PRIVATE_KEY = "5JBTFrjCx11zk8y4N48wjGe3HtqSnZZcgYh7KnDQhbYQZWwVF1e"
	PUBLIC_KEY = "EOS6roDriJXS92hakZ72mrjrCFwrDjitE3FTd9ceuKS85vaZUP2Fg"
)

type errorHandler interface {
	Error(args ...interface{})
}

func newEosWallet(t errorHandler) (*EosWallet, func()) {
	dir := "/galaxy/gc_wallet_test"
	if err := os.Mkdir(path.Join(dir, ".gcwallet"), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			t.Error(err)
		}
	}

	// 设置EOS钱包
	eosCfg := eosconfig.CoinConfig{}
	eosCfg.ClientAPIs = []string{"http://192.168.0.236:8888"}
	eosCfg.CoinType = wi.EOS
	eosCfg.Options = make(map[string]interface{})
	eosCfg.Options["EosparkAPI"] = "https://api.eospark.com/api"
	eosCfg.Options["EosparkAPIKey"] = "3f6a08a55e1b096ce114a3a895e1f2ef"
	var eosDS *eosdb.BoltDatastore
	eosDS, _ = eosdb.Create(dir, "eostest")
	eosCfg.DB = eosDS

	wallet, err := NewEosWallet(eosCfg, dir,"20190101",true)
	//wallet.pool.currentApi().Debug = true
	if err != nil {
		t.Error(err)
	}

	priKey, err := ecc.NewPrivateKey(PRIVATE_KEY)
	if err != nil {
		t.Error(err)
	}

	if err := wallet.ImportKey(priKey, dir, "20190101"); err != nil {
		t.Error(err)
	}

	if err := wallet.ChangeAccount(priKey.PublicKey()); err != nil {
		t.Error(err)
	}

	return wallet, func() {
		wallet.Close()
		os.Remove(viper.GetString("vault.wallet_file"))
		os.RemoveAll(path.Join(dir, ".gcwallet"))
	}
}

func TestEosChainTip(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	height, hash := wallet.ChainTip()

	t.Logf("height is : %d, hash is: %s", height, hash.String())
}

func TestEosGetAccount(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	account, err := wallet.GetAccount("test11111112")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("account is : %+v", account)
}

func TestEosBalance(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	account, err := wallet.db.Accounts().Get(wallet.MasterPublicKey().String())
	if err != nil {
		t.Fatal(err)
	}

	balance := wallet.GetBalance(eos.AN(account.Name))

	t.Logf("balance is : %d", balance)
}

func TestGetTransactions(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	wallet.eospark.Debug = true
	txns, count, err := wallet.GetTransactions( 3, 1, 1, 20)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("count is: %+v", count)
	
	for _, tx := range txns {
		t.Logf("tx is : %+v", tx)
	}
}

func TestGetTransaction(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	txid, err := chainhash.NewHashFromStr("f13c6ddaa22c5a422243c697d7ee5221254ce5e46a928ecd1d7f5fb1577c5488")
	if err != nil {
		t.Fatal(err)
	}
	
	tx, err := wallet.GetTransaction(*txid)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("tx is : %+v", tx)
}

func TestGetTokenList(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	tokenList, err := wallet.GetTokenList()
	if err != nil {
		t.Fatal(err)
	}

	for _, token := range tokenList {
		t.Logf("tx is : %+v", token)
	}
}

func TestNewAccountName(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	accountName, err := wallet.NewAccountName()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("accountName is : %v", accountName)
}

func TestNewAccountByGc(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	key, err := wallet.NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	txid, err := wallet.NewAccountByGC(key.PublicKey(), "")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("accountName is : %v", txid.String())
}

func TestEosSpend(t *testing.T) {
	wallet, cleanup := newEosWallet(t)
	defer cleanup()

	var (
		payerName eos.AccountName = "test11111112"
		payeeName eos.AccountName = "test11111113"
		memo string = "test spend"
		amount float64 = 1
	)

	payerBalance, _ := utils.ToDecimal(wallet.GetBalance(payerName), 4).Float64()
	t.Log("------------")
	t.Logf("payer balance is : %f", payerBalance)
	t.Log("------------")

	beforeBalance, _ := utils.ToDecimal(wallet.GetBalance(payeeName), 4).Float64()
	t.Log("------------")
	t.Logf("payee before balance is : %f", beforeBalance)
	t.Log("------------")

	hash, err := wallet.Spend(payeeName, memo, strconv.FormatFloat(amount, 'f', 4, 64))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("------------")
	t.Logf("hash is : %s", hash.String())
	t.Log("------------")

	afterBalance, _ := utils.ToDecimal(wallet.GetBalance(payeeName), 4).Float64()
	t.Log("------------")
	t.Logf("payee after balance is : %f", afterBalance)
	t.Log("------------")

	if afterBalance - beforeBalance != amount {
		t.Log("------------")
		t.Fatal("amount is not equal")
		t.Log("------------")
	}
}