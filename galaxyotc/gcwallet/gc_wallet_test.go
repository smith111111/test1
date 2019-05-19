package gcwallet

import (
	"os"
	"path"
	"testing"
	"strings"
	"encoding/json"
	"github.com/tyler-smith/go-bip39"
	"github.com/ethereum/go-ethereum/crypto"
	"fmt"
	wallet2 "gcwallet/eoswallet/wallet"
)

const (
	Mnemonic = "entire balance mansion crystal shed offer clock carry outer marine excess crack" // "emotion warm one reform pond law expand return craft veteran maze cute"
	CNMnemonic = "评 逻 劣 尚 敌 纺 土 凯 跟 惊 辩 芳"
	ConfDir = "/galaxy/gc_wallet_test"
	//Password = "20190101"
	Password = "123456"
	// test11111112
	PRIVATE_KEY_1 = "5JBTFrjCx11zk8y4N48wjGe3HtqSnZZcgYh7KnDQhbYQZWwVF1e"
	PUBLIC_KEY_1 = "EOS6roDriJXS92hakZ72mrjrCFwrDjitE3FTd9ceuKS85vaZUP2Fg"
	// test11111113
	PRIVATE_KEY_2 = "5JK9ZQLpyKKuidzoNf4NPTKgArmwb8hS2BVGSWmBmf89LwQg8CY"
	PUBLIC_KEY_3 = "EOS7woS93iF21HofWthqKbvBnuHuns6QDNqdF7GwQYysV99NZdYVJ"

	PRIVATE_KEY = "5KWPQscpPu7t9tskSzp2Wk2pbo6UZERQMr8mmqLsFGGd5o3TwBu"
)

type errorHandler interface {
	Error(args ...interface{})
}

func newGcWallet(t errorHandler) (*GalaxyCoinWallet, func()) {
	if err := os.Mkdir(path.Join(ConfDir, ".gcwallet"), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			t.Error(err)
		}
	}

	gc := New()
	gc.CreateConfig(ConfDir, Mnemonic, Password)
	gc.LoadConfig(ConfDir, Password)

	return gc, func() {
		gc.Close()
		os.RemoveAll(path.Join(ConfDir, ".gcwallet"))
	}
}

func TestGalaxyCoinWallet_NewMnemonic(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	mnemonic := wallet.NewMnemonic()

	t.Logf("mnemonic is: %s", mnemonic)
}

func TestGalaxyCoinWallet_IsMnemonicValid(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	vaild := wallet.IsMnemonicValid(CNMnemonic)

	if vaild {
		seed := bip39.NewSeed(CNMnemonic, "")

		privateKeyECDSA, err := crypto.ToECDSA(seed[:32])
		if err != nil {
			t.Fatal(err)
		}

		address := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey).String()
		fmt.Println(address)
	}

	t.Logf("is memonic: %v", vaild)
}

/* --------------- EOS Wallet Test ---------------*/

func TestGalaxyCoinWallet_EosChainTip(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.EosChainTip()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosBalance(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	_ = wallet.EosImportKey(PRIVATE_KEY, ConfDir, Password)

	response := wallet.EosBalance()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosImportKey(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.EosImportKey(PRIVATE_KEY_1, ConfDir, Password)

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosImportedKeys(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	_ = wallet.EosImportKey(PRIVATE_KEY, ConfDir, Password)
	//_ = wallet.EosImportKey(PRIVATE_KEY_2, ConfDir, Password)

	response := wallet.EosImportedKeys()
	wallet2.GetEosWalletVault(ConfDir, Password)

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosGetTransaction(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.EosGetTransaction("578f2bb176d98254400b0185bf20799c726b863c024c89cdce1f6630c46a5edb")

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosGetTransactions(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	// 要使用正式网的私钥
	//_ = wallet.EosImportKey("???", ConfDir, Password)

	response := wallet.EosGetTransactions(3, 1, 1, 20)

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosSpend(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	_ = wallet.EosImportKey(PRIVATE_KEY_1, ConfDir, Password)

	response := wallet.EosSpend("test11111113", "123456", "1")

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosNewAccountName(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response :=  wallet.EosNewAccountName()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EosNewAccountByGC(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	key, err := wallet.eosWallet.NewKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("private key is %s, public key is %s", key.String(), key.PublicKey().String())

	response :=  wallet.EosNewAccountByGC(key.PublicKey().String(), "jonhoychan11")

	t.Logf("response is: %s", response)
}

/*--------------- ETH Wallet Test ---------------*/
func TestGalaxyCoinWallet_EthChainTip(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.EthChainTip()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EthBalance(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.EthBalance()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_EthTransactions(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()
	wallet.Start()

	response := wallet.EthTransactions( 3, 1, 1, 20)

	t.Logf("response is: %s", response)
}

/* --------------- GC Wallet Test ---------------*/

func TestGalaxyCoinWallet_GcBalance(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()

	response := wallet.GcBalance()

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_GcTransactions(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()
	wallet.Start()

	response := wallet.GcTransactions( 1, 1, 1, 20)

	t.Logf("response is: %s", response)
}

/* --------------- Multi Wallet Test ---------------*/

func TestGalaxyCoinWallet_MultiTransactions(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()
	wallet.Start()

	response := wallet.MultiTransactions("BCH", 3, 1, 1, 20)

	t.Logf("response is: %s", response)

	txResposne := wallet.MultiGetTransaction("BCH", "a2fd14c1f49929cc985657a520c5388b5361ea74be42bbfb6b41af0458772736")

	t.Logf("txResposne is: %s", txResposne)
}

func TestGalaxyCoinWallet_MultiSpend(t *testing.T) {
	wallet, cleanup := newGcWallet(t)
	defer cleanup()
	wallet.Start()

	response := wallet.MultiSpend("BCH", "10000", "qq9fkrqgl3pc4085m9dpgcevjrq4p8cxxq3dkyn8v5", "normal")

	t.Logf("response is: %s", response)
}

func TestGalaxyCoinWallet_JsonDecode(t *testing.T) {
	jsonStream := `
		{"txid":{"result":"d3a86b560d73010b5f4619f39b2660a71cd6c21eda2db61b5c93361e54ec75c0","error":null,"id":63936}}
	`

	type Response struct {
		Txid string `json:"txid"`
	}

	rs := new(Response)
	if err := json.NewDecoder(strings.NewReader(jsonStream)).Decode(rs); err != nil {
		type Txid struct {
			Result string `json:"result"`
		}
		type BCHResponse struct {
			Txid Txid `json:"txid"`
		}
		rs := new(BCHResponse)
		if err = json.NewDecoder(strings.NewReader(jsonStream)).Decode(rs); err != nil {
			t.Fatal(err)
		}

		t.Logf("txid is: %s", rs.Txid)
	}
}