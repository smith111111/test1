package wallet

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"gcwallet/go-ethwallet/boltdb"
	"io/ioutil"
	"math/big"
	"path"
	"runtime"
	"time"

	"gcwallet/go-ethwallet/config"
	wi "gcwallet/eth-wallet-interface"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	hd "github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v2"

	"gcwallet/go-ethwallet/util"
	"gcwallet/go-ethwallet/etherscan"
	"strings"
	"strconv"
)

// EthConfiguration - used for eth specific configuration
type EthConfiguration struct {
	RopstenPPAddress string `yaml:"ROPSTEN_PPv2_ADDRESS"`
	RegistryAddress  string `yaml:"ROPSTEN_REGISTRY"`
}

// EthRedeemScript - used to represent redeem script for eth wallet
// <uniqueId: 20><threshold:1><timeoutHours:4><buyer:20><seller:20>
// <moderator:20><multisigAddress:20><tokenAddress:20>
type EthRedeemScript struct {
	TxnID           common.Address
	Threshold       uint8
	Timeout         uint32
	Buyer           common.Address
	Seller          common.Address
	Moderator       common.Address
	MultisigAddress common.Address
	TokenAddress    common.Address
}

// SerializeEthScript - used to serialize eth redeem script
func SerializeEthScript(scrpt EthRedeemScript) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(scrpt)
	return b.Bytes(), err
}

// DeserializeEthScript - used to deserialize eth redeem script
func DeserializeEthScript(b []byte) (EthRedeemScript, error) {
	scrpt := EthRedeemScript{}
	buf := bytes.NewBuffer(b)
	d := gob.NewDecoder(buf)
	err := d.Decode(&scrpt)
	return scrpt, err
}

// GenScriptHash - used to generate script hash for eth as per
// escrow smart contract
func GenScriptHash(script EthRedeemScript) ([32]byte, string, error) {
	ahash := sha3.NewKeccak256()
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, script.Timeout)
	arr := append(script.TxnID.Bytes(), append([]byte{script.Threshold},
		append(a[:], append(script.Buyer.Bytes(),
			append(script.Seller.Bytes(), append(script.Moderator.Bytes(),
				append(script.MultisigAddress.Bytes())...)...)...)...)...)...)
	ahash.Write(arr)
	var retHash [32]byte

	copy(retHash[:], ahash.Sum(nil)[:])
	ahashStr := hexutil.Encode(retHash[:])

	return retHash, ahashStr, nil
}

// EthereumWallet is the wallet implementation for ethereum
type EthereumWallet struct {
	client        *EthClient
	account       *Account
	address       *EthAddress
	service       *Service
	registry      *Registry
	ppsct         *Escrow
	db            wi.Datastore
	exchangeRates wi.ExchangeRates
	listeners     []func(wi.TransactionCallback)
}

// NewEthereumWalletWithKeyfile will return a reference to the Eth Wallet
func NewEthereumWalletWithKeyfile(url, keyFile, passwd string) *EthereumWallet {
	client, err := NewEthClient(url)
	if err != nil {
		log.Fatalf("error initializing wallet: %v", err)
	}
	var myAccount *Account
	myAccount, err = NewAccountFromKeyfile(keyFile, passwd)
	if err != nil {
		log.Fatalf("key file validation failed: %s", err.Error())
	}
	addr := myAccount.Address()

	_, filename, _, _ := runtime.Caller(0)
	conf, err := ioutil.ReadFile(path.Join(path.Dir(filename), "../configuration.yaml"))

	if err != nil {
		log.Fatalf("ethereum config not found: %s", err.Error())
	}
	ethConfig := EthConfiguration{}
	err = yaml.Unmarshal(conf, &ethConfig)
	if err != nil {
		log.Fatalf("ethereum config not valid: %s", err.Error())
	}

	reg, err := NewRegistry(common.HexToAddress(ethConfig.RegistryAddress), client)
	if err != nil {
		log.Fatalf("error initilaizing contract failed: %s", err.Error())
	}

	//reg.GetVersionDetails()

	//smtct, err := NewWallet(common.HexToAddress(ethConfig.RopstenPPAddress), client)
	//if err != nil {
	//	log.Fatalf("error initilaizing contract failed: %s", err.Error())
	//}

	return &EthereumWallet{client, myAccount, &EthAddress{&addr}, &Service{}, reg, nil, nil, nil, []func(wi.TransactionCallback){}}
}

// NewEthereumWallet will return a reference to the Eth Wallet
func NewEthereumWallet(cfg config.CoinConfig, mnemonic string, proxy proxy.Dialer) (*EthereumWallet, error) {
	// ZJ:
	client, err := NewEthClient(cfg.ClientAPIs[0])
	//client, err := NewEthClient(cfg.ClientAPI.String() + "/" + InfuraAPIKey)
	if err != nil {
		log.Errorf("error initializing wallet: %v", err)
		return nil, err
	}
	var myAccount *Account
	/*
		seed := bip39.NewSeed(mnemonic, "")

		privateKeyECDSA, err := crypto.ToECDSA(seed)
		if err != nil {
			return nil
		}

		key := &keystore.Key{
			Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
			PrivateKey: privateKeyECDSA,
		}

		myAccount = &Account{key}
	*/
	myAccount, err = NewAccountFromMnemonic(mnemonic, "")
	if err != nil {
		log.Errorf("mnemonic based pk generation failed: %s", err.Error())
		return nil, err
	}
	addr := myAccount.Address()

	etherscanApiUrl, ok := cfg.Options["EtherscanAPI"]
	if !ok {
		return nil, errors.New("can not find etherscan api url")
	}
	etherscanApiKey, ok := cfg.Options["EtherscanAPIKey"]
	if !ok {
		return nil, errors.New("can not find etherscan api key")
	}

	// new ehtersan api
	ea := etherscan.New(etherscanApiUrl.(string), etherscanApiKey.(string))
	if err != nil {
		return nil, err
	}

	ws := NewWalletService(cfg.DB, client.Client, cfg.CoinType, addr, ea, "")

	ethConfig := EthConfiguration{}

	regAddr, ok := cfg.Options["RegistryAddress"];
	if !ok {
		log.Errorf("ethereum registry not found: %s", err.Error())
		return nil, err
	}

	ethConfig.RegistryAddress = regAddr.(string)

	/*
		_, filename, _, _ := runtime.Caller(0)
		conf, err := ioutil.ReadFile(path.Join(path.Dir(filename), "../configuration.yaml"))

		if err != nil {
			log.Errorf("ethereum config not found: %s", err.Error())
			return nil, err
		}

		err = yaml.Unmarshal(conf, &ethConfig)
		if err != nil {
			log.Errorf("ethereum config not valid: %s", err.Error())
			return nil, err
		}
	*/

	reg, err := NewRegistry(common.HexToAddress(ethConfig.RegistryAddress), client)
	if err != nil {
		log.Errorf("error initilaizing contract failed: %s", err.Error())
		return nil, err
	}

	//reg.GetVersionDetails()

	//smtct, err := NewWallet(common.HexToAddress(ethConfig.RopstenPPAddress), client)
	//if err != nil {
	//	log.Fatalf("error initilaizing contract failed: %s", err.Error())
	//}

	er := NewEthereumPriceFetcher(proxy)

	return &EthereumWallet{client, myAccount, &EthAddress{&addr}, ws, reg, nil, cfg.DB, er, []func(wi.TransactionCallback){}}, nil
}

// Params - return nil to comply
func (wallet *EthereumWallet) Params() *chaincfg.Params {
	return nil
}

// GetBalance returns the balance for the wallet
func (wallet *EthereumWallet) GetBalance() (*big.Int, error) {
	return wallet.client.GetBalance(wallet.account.Address())
}

// GetUnconfirmedBalance returns the unconfirmed balance for the wallet
func (wallet *EthereumWallet) GetUnconfirmedBalance() (*big.Int, error) {
	return wallet.client.GetUnconfirmedBalance(wallet.account.Address())
}

// Transfer will transfer the amount from this wallet to the spec address
func (wallet *EthereumWallet) Transfer(to string, value *big.Int) (common.Hash, error) {
	toAddress := common.HexToAddress(to)
	return wallet.client.Transfer(wallet.account, toAddress, value)
}

// Start will start the wallet daemon
func (wallet *EthereumWallet) Start() {
	// daemonize the wallet
	wallet.service.Start()
}

// CurrencyCode returns ETH
func (wallet *EthereumWallet) CurrencyCode() string {
	return "ETH"
}

// IsDust Check if this amount is considered dust - 10000 wei
func (wallet *EthereumWallet) IsDust(amount int64) bool {
	return amount < 10000
}

// MasterPrivateKey - Get the master private key
func (wallet *EthereumWallet) MasterPrivateKey() *hd.ExtendedKey {
	return hd.NewExtendedKey([]byte{0x00, 0x00, 0x00, 0x00}, wallet.account.privateKey.D.Bytes(),
		wallet.account.address.Bytes(), wallet.account.address.Bytes(), 0, 0, true)
}

// MasterPublicKey - Get the master public key
func (wallet *EthereumWallet) MasterPublicKey() *hd.ExtendedKey {
	publicKey := wallet.account.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	return hd.NewExtendedKey([]byte{0x00, 0x00, 0x00, 0x00}, publicKeyBytes,
		wallet.account.address.Bytes(), wallet.account.address.Bytes(), 0, 0, false)
}

// ChildKey Generate a child key using the given chaincode. The key is used in multisig transactions.
// For most implementations this should just be child key 0.
func (wallet *EthereumWallet) ChildKey(keyBytes []byte, chaincode []byte, isPrivateKey bool) (*hd.ExtendedKey, error) {
	if isPrivateKey {
		return wallet.MasterPrivateKey(), nil
	}
	return wallet.MasterPublicKey(), nil
}

// CurrentAddress - Get the current address for the given purpose
func (wallet *EthereumWallet) CurrentAddress(purpose wi.KeyPurpose) btcutil.Address {
	return *wallet.address
}

// NewAddress - Returns a fresh address that has never been returned by this function
func (wallet *EthereumWallet) NewAddress(purpose wi.KeyPurpose) btcutil.Address {
	return *wallet.address
}

// DecodeAddress - Parse the address string and return an address interface
func (wallet *EthereumWallet) DecodeAddress(addr string) (btcutil.Address, error) {
	ethAddr := common.HexToAddress(addr)
	if wallet.HasKey(EthAddress{&ethAddr}) {
		return *wallet.address, nil
	}
	return EthAddress{}, errors.New("invalid or unknown address")
}

// ScriptToAddress - ?
func (wallet *EthereumWallet) ScriptToAddress(script []byte) (btcutil.Address, error) {
	return wallet.address, nil
}

// HasKey - Returns if the wallet has the key for the given address
func (wallet *EthereumWallet) HasKey(addr btcutil.Address) bool {
	if !util.IsValidAddress(addr.String()) {
		return false
	}
	return wallet.account.Address().String() == addr.String()
}

// Balance - Get the confirmed and unconfirmed balances
func (wallet *EthereumWallet) Balance() (confirmed, unconfirmed int64) {
	var balance, ucbalance int64
	bal, err := wallet.GetBalance()
	if err == nil {
		balance = bal.Int64()
	}
	ucbal, err := wallet.GetUnconfirmedBalance()
	if err == nil {
		ucbalance = ucbal.Int64()
	}
	return balance, ucbalance
}

func (wallet *EthereumWallet) BalanceBigInt() (confirmed, unconfirmed big.Int) {
	balance := big.NewInt(0)
	ucbalance := big.NewInt(0)
	bal, err := wallet.GetBalance()
	if err == nil {
		balance = bal
	}
	ucbal, err := wallet.GetUnconfirmedBalance()
	if err == nil {
		ucbalance = ucbal
	}
	return *balance, *ucbalance
}

// Transactions - Returns a list of transactions for this wallet
func (wallet *EthereumWallet) Transactions(transactionType, sort, offset, limit int) ([]wi.Txn, int32, error) {
	height, _ := wallet.ChainTip()
	txns, count, err := wallet.db.Txns().GetAllByLimit(wallet.account.Address().String(), transactionType, sort, offset, limit)
	if err != nil {
		return txns, count, err
	}
	for i, tx := range txns {
		confirmations := int32(height) - tx.Height + 1
		if tx.Height <= 0 {
			confirmations = tx.Height
		}

		tx.Confirmations = int64(confirmations)

		var value string
		// 根据转账接收者判断该交易是支出还是收入
		if strings.ToLower(tx.To.String()) != strings.ToLower(wallet.account.Address().String()) {
			amount, _ := strconv.ParseInt(tx.Value, 10, 64)
			if amount != 0 {
				value = fmt.Sprintf("-%s", tx.Value)
			} else {
				value = tx.Value
			}
		}
		tx.Value = value
		txns[i] = tx
	}
	return txns, count, nil
}

// GetTransaction - Get info on a specific transaction
func (wallet *EthereumWallet) GetTransaction(txid chainhash.Hash) (wi.Txn, error) {
	height, _ := wallet.ChainTip()
	txn, err := wallet.db.Txns().Get(txid.String())
	if err != nil {
		return txn, err
	}
	confirmations := int32(height) - txn.Height + 1
	if txn.Height <= 0 {
		confirmations = txn.Height
	}

	txn.Confirmations = int64(confirmations)

	var value string
	// 根据转账接收者判断该交易是支出还是收入
	if strings.ToLower(txn.To.String()) != strings.ToLower(wallet.account.Address().String()) {
		amount, _ := strconv.ParseInt(txn.Value, 10, 64)
		if amount != 0 {
			value = fmt.Sprintf("-%s", txn.Value)
		} else {
			value = txn.Value
		}
	}
	txn.Value = value

	return txn, nil
}

// ChainTip - Get the height and best hash of the blockchain
func (wallet *EthereumWallet) ChainTip() (uint32, chainhash.Hash) {
	num, hash, err := wallet.client.GetLatestBlock()
	h, _ := chainhash.NewHashFromStr("")
	if err != nil {
		return 0, *h
	}
	h, _ = chainhash.NewHashFromStr(hash[2:])
	return num, *h
}

// GetFeePerByte - Get the current fee per byte
func (wallet *EthereumWallet) GetFeePerByte(feeLevel wi.FeeLevel) uint64 {
	return 0
}

// Spend - Send ether to an external wallet
func (wallet *EthereumWallet) Spend(amount string, addr btcutil.Address, feeLevel wi.FeeLevel, referenceID string) (*chainhash.Hash, error) {

	var hash common.Hash
	var h *chainhash.Hash
	var err error

	//twoMinutes, _ := time.ParseDuration("2m")

	// ZJ:
	isScript := false
	redeemScript := []byte{}
	if wallet.db != nil{
		// check if the addr is a multisig addr
		scripts, err := wallet.db.WatchedScripts().GetAll()
		if err != nil {
			return nil, err
		}

		key := []byte(addr.String())
		for _, script := range scripts {
			if bytes.Equal(key, script[:common.AddressLength]) {
				isScript = true
				redeemScript = script[common.AddressLength:]
				break
			}
		}
	}

	amountBig, _ := new(big.Int).SetString(amount, 10)
	if isScript {
		ethScript, err := DeserializeEthScript(redeemScript)
		if err != nil {
			return nil, err
		}
		hash, err = wallet.callAddTransaction(ethScript, amountBig)
	} else {
		hash, err = wallet.Transfer(addr.String(), amountBig)
	}

	if err != nil {
		return nil, err
	}
	start := time.Now()
	flag := false
	var rcpt *types.Receipt
	for !flag {
		rcpt, err = wallet.client.TransactionReceipt(context.Background(), hash)
		if rcpt != nil {
			flag = true
		}
		if time.Since(start).Seconds() > 120 {
			flag = true
		}
	}
	if rcpt != nil {
		// good. so the txn has been processed but we have to account for failed
		// but valid txn like some contract condition causing revert
		if rcpt.Status > 0 {
			// all good to update order state
			go wallet.callListeners(wallet.createTxnCallback(hash.String(), referenceID, amount, time.Now()))
		} else {
			// there was some error processing this txn
			return nil, errors.New(fmt.Sprintf("problem processing this transaction, txid:%s",hash.String()[2:]))
		}
	}

	if err == nil {
		// ZJ:
		//h, err = chainhash.NewHashFromStr(hash.String())
		h, err = chainhash.NewHashFromStr(hash.String()[2:])
	}
	return h, err
}

func (wallet *EthereumWallet) createTxnCallback(txID, orderID string, value string, bTime time.Time) wi.TransactionCallback {
	output := wi.TransactionOutput{
		Address: wallet.address,
		Value:   value,
		Index:   1,
		OrderID: orderID,
	}

	return wi.TransactionCallback{
		Txid:      txID,
		Outputs:   []wi.TransactionOutput{output},
		Inputs:    []wi.TransactionInput{},
		Height:    1,
		Timestamp: time.Now(),
		Value:     value,
		WatchOnly: false,
		BlockTime: bTime,
	}
}

func (wallet *EthereumWallet) callListeners(txnCB wi.TransactionCallback) {
	for _, l := range wallet.listeners {
		go l(txnCB)
	}
}

// BumpFee - Bump the fee for the given transaction
func (wallet *EthereumWallet) BumpFee(txid chainhash.Hash) (*chainhash.Hash, error) {
	return chainhash.NewHashFromStr(txid.String())
}

// EstimateFee - Calculates the estimated size of the transaction and returns the total fee for the given feePerByte
func (wallet *EthereumWallet) EstimateFee(ins []wi.TransactionInput, outs []wi.TransactionOutput, feePerByte uint64) uint64 {
	sum := big.NewInt(0)
	for _, out := range outs {
		valueBig, _ := new(big.Int).SetString(out.Value, 10)
		gas, err := wallet.client.EstimateTxnGas(wallet.account.Address(),
			common.HexToAddress(out.Address.String()), valueBig)
		if err != nil {
			return sum.Uint64()
		}
		sum.Add(sum, gas)
	}
	return sum.Uint64()
}

// EstimateSpendFee - Build a spend transaction for the amount and return the transaction fee
func (wallet *EthereumWallet) EstimateSpendFee(amount int64, feeLevel wi.FeeLevel) (uint64, error) {
	gas, err := wallet.client.EstimateGasSpend(wallet.account.Address(), big.NewInt(amount))
	return gas.Uint64(), err
}

// SweepAddress - Build and broadcast a transaction that sweeps all coins from an address. If it is a p2sh multisig, the redeemScript must be included
func (wallet *EthereumWallet) SweepAddress(utxos []wi.TransactionInput, address *btcutil.Address, key *hd.ExtendedKey, redeemScript *[]byte, feeLevel wi.FeeLevel) (*chainhash.Hash, error) {
	return chainhash.NewHashFromStr("")
}

// ExchangeRates - return the exchangerates
func (wallet *EthereumWallet) ExchangeRates() wi.ExchangeRates {
	return wallet.exchangeRates
}

func (wallet *EthereumWallet) callAddTransaction(script EthRedeemScript, value *big.Int) (common.Hash, error) {

	h := common.BigToHash(big.NewInt(0))

	// call registry to get the deployed address for the escrow ct
	fromAddress := wallet.account.Address()
	nonce, err := wallet.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := wallet.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(wallet.account.privateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = value      // in wei
	auth.GasLimit = 4000000 // in units
	auth.GasPrice = gasPrice

	//redeemScript, err := SerializeEthScript(script)
	//if err != nil {
	//	return h, err
	//}

	shash, _, err := GenScriptHash(script)
	if err != nil {
		return h, err
	}

	smtct, err := NewEscrow(script.MultisigAddress, wallet.client)
	if err != nil {
		log.Fatalf("error initilaizing contract failed: %s", err.Error())
	}

	///smtct.CalculateRedeemScriptHash()

	var tx *types.Transaction

	//if script.Threshold == 1 {
	tx, err = smtct.AddTransaction(auth, script.Buyer, script.Seller,
		script.Moderator, script.Threshold,
		script.Timeout, shash, script.TxnID)
	//} else {
	//	tx, err = smtct.AddTransaction(auth, script.Buyer, script.Seller,
	//		script.Moderator, script.Threshold,
	//		script.Timeout, shash, script.TxnID)
	//}

	if err == nil {
		h = tx.Hash()
	}

	return h, err

}

// GenerateMultisigScript - Generate a multisig script from public keys. If a timeout is included the returned script should be a timelocked escrow which releases using the timeoutKey.
func (wallet *EthereumWallet) GenerateMultisigScript(keys []hd.ExtendedKey, threshold int, timeout time.Duration, timeoutKey *hd.ExtendedKey) (btcutil.Address, []byte, error) {
	if uint32(timeout.Hours()) > 0 && timeoutKey == nil {
		return nil, nil, errors.New("timeout key must be non nil when using an escrow timeout")
	}

	if len(keys) < threshold {
		return nil, nil, fmt.Errorf("unable to generate multisig script with "+
			"%d required signatures when there are only %d public "+
			"keys available", threshold, len(keys))
	}

	if len(keys) < 2 {
		return nil, nil, fmt.Errorf("unable to generate multisig script with "+
			"%d required signatures when there are only %d public "+
			"keys available", threshold, len(keys))
	}

	var ecKeys []common.Address
	for _, key := range keys {
		ecKey, err := key.ECPubKey()
		if err != nil {
			return nil, nil, err
		}
		ecKeys = append(ecKeys, common.BytesToAddress(ecKey.SerializeUncompressed()))
	}

	ver, err := wallet.registry.GetRecommendedVersion(nil, "escrow")
	if err != nil {
		log.Fatal(err)
	}

	if util.IsZeroAddress(ver.Implementation) {
		return nil, nil, errors.New("no escrow contract available")
	}

	builder := EthRedeemScript{}

	builder.TxnID = common.BytesToAddress(util.ExtractChaincode(&keys[0]))
	builder.Timeout = uint32(timeout.Hours())
	builder.Threshold = uint8(threshold)
	builder.Buyer = ecKeys[0]
	builder.Seller = ecKeys[1]
	builder.MultisigAddress = ver.Implementation

	if threshold > 1 {
		builder.Moderator = ecKeys[2]
	}
	switch threshold {
	case 1:
		{
			// Seller is offline
		}
	case 2:
		{
			// Moderated payment
		}
	default:
		{
			// handle this
		}
	}

	redeemScript, err := SerializeEthScript(builder)
	if err != nil {
		return nil, nil, err
	}

	hash := sha3.NewKeccak256()
	hash.Write(redeemScript)
	addr := common.HexToAddress(hexutil.Encode(hash.Sum(nil)[:]))
	retAddr := EthAddress{&addr}

	scriptKey := append(addr.Bytes(), redeemScript...)
	wallet.db.WatchedScripts().Put(scriptKey)

	return retAddr, redeemScript, nil
}

// CreateMultisigSignature - Create a signature for a multisig transaction
func (wallet *EthereumWallet) CreateMultisigSignature(ins []wi.TransactionInput, outs []wi.TransactionOutput, key *hd.ExtendedKey, redeemScript []byte, feePerByte uint64) ([]wi.Signature, error) {

	var sigs []wi.Signature

	payables := make(map[string]*big.Int)
	for _, out := range outs {
		val, _ := new(big.Int).SetString(out.Value, 10)
		if val.Cmp(big.NewInt(0)) <= 0 {
			continue
		}

		if p, ok := payables[out.Address.String()]; ok {
			sum := big.NewInt(0)
			sum.Add(val, p)
			payables[out.Address.String()] = sum
		} else {
			payables[out.Address.String()] = val
		}
	}

	//destinations := []common.Address{}
	//amounts := []*big.Int{}

	destArr := []byte{}
	//destStr := ""
	amountArr := []byte{}
	//amountStr := ""

	//spew.Dump(payables)

	for k, v := range payables {
		addr := common.HexToAddress(k)
		sample := [32]byte{}
		sampleDest := [32]byte{}
		copy(sampleDest[12:], addr.Bytes())
		a := make([]byte, 8)
		binary.BigEndian.PutUint64(a, v.Uint64())

		copy(sample[24:], a)
		//destinations = append(destinations, addr)
		//amounts = append(amounts, v)
		//addrStr := fmt.Sprintf("%064s", addr.String())
		//destStr = destStr + addrStr
		destArr = append(destArr, sampleDest[:]...)
		amountArr = append(amountArr, sample[:]...)
		//amnt := fmt.Sprintf("%064s", fmt.Sprintf("%x", v.Int64()))
		//amountStr = amountStr + amnt
	}

	//fmt.Println("destarr     : ", destArr)
	//fmt.Println("amountArr   : ", amountArr)

	rScript, err := DeserializeEthScript(redeemScript)
	if err != nil {
		return nil, err
	}

	//spew.Dump(rScript)

	shash, _, err := GenScriptHash(rScript)
	if err != nil {
		return nil, err
	}

	var txHash [32]byte
	var payloadHash [32]byte

	/*
				// Follows ERC191 signature scheme: https://github.com/ethereum/EIPs/issues/191
		        bytes32 txHash = keccak256(
		            abi.encodePacked(
		                "\x19Ethereum Signed Message:\n32",
		                keccak256(
		                    abi.encodePacked(
		                        byte(0x19),
		                        byte(0),
		                        this,
		                        destinations,
		                        amounts,
		                        scriptHash
		                    )
		                )
		            )
		        );

	*/

	payload := []byte{byte(0x19), byte(0)}
	payload = append(payload, rScript.MultisigAddress.Bytes()...)
	payload = append(payload, destArr...)
	payload = append(payload, amountArr...)
	payload = append(payload, shash[:]...)

	//script.MultisigAddress.String()[2:]

	//payloadStr := "0x19" + "00" + rScript.MultisigAddress.String()[2:] + destStr + amountStr +
	//	hashStr[2:]
	//payloadStr = payloadStr + ""

	//pHash := crypto.Keccak256([]byte(payloadStr))
	pHash := crypto.Keccak256(payload)
	copy(payloadHash[:], pHash)

	txData := []byte{byte(0x19)}
	txData = append(txData, []byte("Ethereum Signed Message:\n32")...)
	//txData = append(txData, byte(32))
	txData = append(txData, payloadHash[:]...)
	txnHash := crypto.Keccak256(txData)
	//fmt.Println("txnHash        : ", hexutil.Encode(txnHash))
	//fmt.Println("phash          : ", hexutil.Encode(payloadHash[:]))
	copy(txHash[:], txnHash)

	sig, err := crypto.Sign(txHash[:], wallet.account.privateKey)
	if err != nil {
		log.Errorf("error signing in createmultisig : %v", err)
	}
	sigs = append(sigs, wi.Signature{InputIndex: 1, Signature: sig})

	return sigs, err
}

// Multisign - Combine signatures and optionally broadcast
func (wallet *EthereumWallet) Multisign(ins []wi.TransactionInput, outs []wi.TransactionOutput, sigs1 []wi.Signature, sigs2 []wi.Signature, redeemScript []byte, feePerByte uint64, broadcast bool) ([]byte, error) {

	//var buf bytes.Buffer

	payables := make(map[string]*big.Int)
	for _, out := range outs {
		val, _ := new(big.Int).SetString(out.Value, 10)
		if val.Cmp(big.NewInt(0)) <= 0 {
			continue
		}

		if p, ok := payables[out.Address.String()]; ok {
			sum := big.NewInt(0)
			sum.Add(val, p)
			payables[out.Address.String()] = sum
		} else {
			payables[out.Address.String()] = val
		}
	}

	rSlice := [][32]byte{} //, 2)
	sSlice := [][32]byte{} //, 2)
	vSlice := []uint8{}    //, 2)

	r := [32]byte{}
	s := [32]byte{}
	v := uint8(0)

	if len(sigs1[0].Signature) > 0 {
		r, s, v = util.SigRSV(sigs1[0].Signature)
		rSlice = append(rSlice, r)
		sSlice = append(sSlice, s)
		vSlice = append(vSlice, v)
	}

	r = [32]byte{}
	s = [32]byte{}
	v = uint8(0)

	if len(sigs2[0].Signature) > 0 {
		r, s, v = util.SigRSV(sigs2[0].Signature)
		rSlice = append(rSlice, r)
		sSlice = append(sSlice, s)
		vSlice = append(vSlice, v)
	}

	rScript, err := DeserializeEthScript(redeemScript)
	if err != nil {
		return nil, err
	}

	shash, _, err := GenScriptHash(rScript)
	if err != nil {
		return nil, err
	}

	smtct, err := NewEscrow(rScript.MultisigAddress, wallet.client)
	if err != nil {
		log.Fatalf("error initilaizing contract failed: %s", err.Error())
	}

	destinations := []common.Address{}
	amounts := []*big.Int{}

	for k, v := range payables {
		destinations = append(destinations, common.HexToAddress(k))
		amounts = append(amounts, v)
	}

	fromAddress := wallet.account.Address()
	nonce, err := wallet.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasPrice, err := wallet.client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	auth := bind.NewKeyedTransactor(wallet.account.privateKey)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = 4000000    // in units
	auth.GasPrice = gasPrice

	var tx *types.Transaction

	tx, err = smtct.Execute(auth, vSlice, rSlice, sSlice, shash, destinations, amounts)

	//fmt.Println(tx)
	//fmt.Println(err)

	if err != nil {
		return nil, err
	}

	ret, err := tx.MarshalJSON()

	return ret, err
}

// AddWatchedAddress - Add a script to the wallet and get notifications back when coins are received or spent from it
func (wallet *EthereumWallet) AddWatchedAddress(address btcutil.Address) error {
	// the reason eth wallet cannot use this as of now is because only the address
	// is insufficient, the redeemScript is also required
	return nil
}

// AddTransactionListener - add a txn listener
func (wallet *EthereumWallet) AddTransactionListener(callback func(wi.TransactionCallback)) {
	// add incoming txn listener using service
}

// ReSyncBlockchain - Use this to re-download merkle blocks in case of missed transactions
func (wallet *EthereumWallet) ReSyncBlockchain(fromTime time.Time) {
	// use service here
}

// GetConfirmations - Return the number of confirmations and the height for a transaction
func (wallet *EthereumWallet) GetConfirmations(txid chainhash.Hash) (confirms, atHeight uint32, err error) {
	return 0, 0, nil
}

// Close will stop the wallet daemon
func (wallet *EthereumWallet) Close() {
	if wallet.db != nil{
		db := wallet.db.(*boltdb.BoltDatastore)
		db.DB.Close()
	}
}

// CreateAddress - used to generate a new address
func (wallet *EthereumWallet) CreateAddress() (common.Address, error) {
	fromAddress := wallet.account.Address()
	nonce, err := wallet.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
	}
	addr := crypto.CreateAddress(fromAddress, nonce)
	//fmt.Println("Addr : ", addr.String())
	return addr, err
}

// GenWallet creates a wallet
func GenWallet() {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println(hexutil.Encode(privateKeyBytes)[2:]) // fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println(hexutil.Encode(publicKeyBytes)[4:]) // 9a7df67f79246283fdc93af76d4f8cdd62c4886e8cd870944e817dd0b97934fdd7719d0810951e03418205868a5c1b40b192451367f28e0088dd75e15de40c05

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println(address) // 0x96216849c49358B10257cb55b28eA603c874b05E

	hash := sha3.NewKeccak256()
	hash.Write(publicKeyBytes[1:])
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:])) // 0x96216849c49358b10257cb55b28ea603c874b05e

	fmt.Println(util.IsValidAddress(address))

}

// GenDefaultKeyStore will generate a default keystore
func GenDefaultKeyStore(passwd string) (*Account, error) {
	ks := keystore.NewKeyStore("./", keystore.StandardScryptN, keystore.StandardScryptP)

	account, err := ks.NewAccount(passwd)
	if err != nil {
		return nil, err
	}
	return NewAccountFromKeyfile(account.URL.Path, passwd)
}

// ZJ:
func (wallet *EthereumWallet) SuggestGasPrice() (*big.Int, error) {
	return wallet.client.SuggestGasPrice(context.Background())
}