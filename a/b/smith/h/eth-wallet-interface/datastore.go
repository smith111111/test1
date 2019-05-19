package wallet

import (
	"bytes"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"time"
	"github.com/ethereum/go-ethereum/common"
)

type Coin interface {
	String() string
	CurrencyCode() string
}

type CoinType uint32

const (
	Bitcoin     CoinType = 0
	Litecoin             = 1
	Zcash                = 133
	BitcoinCash          = 145
	Ethereum             = 60

	TestnetBitcoin       = 1000000
	TestnetLitecoin      = 1000001
	TestnetZcash         = 1000133
	TestnetBitcoinCash   = 1000145
	TestnetEthereum      = 1000060
)

func (c *CoinType) String() string {
	switch *c {
	case Bitcoin:
		return "Bitcoin"
	case BitcoinCash:
		return "Bitcoin Cash"
	case Zcash:
		return "Zcash"
	case Litecoin:
		return "Litecoin"
	case Ethereum:
		return "Ethereum"
	case TestnetBitcoin:
		return "Testnet Bitcoin"
	case TestnetBitcoinCash:
		return "Testnet Bitcoin Cash"
	case TestnetZcash:
		return "Testnet Zcash"
	case TestnetLitecoin:
		return "Testnet Litecoin"
	case TestnetEthereum:
		return "Testnet Ethereum"
	default:
		return ""
	}
}

func (c *CoinType) CurrencyCode() string {
	switch *c {
	case Bitcoin:
		return "BTC"
	case BitcoinCash:
		return "BCH"
	case Zcash:
		return "ZEC"
	case Litecoin:
		return "LTC"
	case Ethereum:
		return "ETH"
	case TestnetBitcoin:
		return "TBTC"
	case TestnetBitcoinCash:
		return "TBCH"
	case TestnetZcash:
		return "TZEC"
	case TestnetLitecoin:
		return "TLTC"
	case TestnetEthereum:
		return "TETH"
	default:
		return ""
	}
}

type Datastore interface {
	Utxos() Utxos
	Stxos() Stxos
	Txns() Txns
	Keys() Keys
	WatchedScripts() WatchedScripts
}

type Utxos interface {
	// Put a utxo to the database
	Put(utxo Utxo) error

	// Fetch all utxos from the db
	GetAll() ([]Utxo, error)

	// Make a utxo unspendable
	SetWatchOnly(utxo Utxo) error

	// Delete a utxo from the db
	Delete(utxo Utxo) error
}

type Stxos interface {
	// Put a stxo to the database
	Put(stxo Stxo) error

	// Fetch all stxos from the db
	GetAll() ([]Stxo, error)

	// Delete a stxo from the db
	Delete(stxo Stxo) error
}

type Txns interface {
	// Put a new transaction to the database
	Put(txid, value, from, to, gas, status, contract string, height int32, timestamp int64, input string) error

	// Fetch a tx and it's metadata given a hash
	Get(txid string) (Txn, error)

	// Fetch all transactions from the db
	GetAll() ([]Txn, error)

	// Fetch all transactions by limit from the db
	GetAllByLimit(address string, transactionType, sort, offset, limit int) ([]Txn, int32, error)

	// Update the height of a transaction
	UpdateHeight(txid string, height int32, timestamp int64) error

	// Delete a transactions from the db
	Delete(txid string) error
}

// Keys provides a database interface for the wallet to save key material, track
// used keys, and manage the look ahead window.
type Keys interface {
	// Put a bip32 key to the database
	Put(hash160 []byte, keyPath KeyPath) error

	// Import a loose private key not part of the keychain
	ImportKey(scriptAddress []byte, key *btcec.PrivateKey) error

	// Mark the script as used
	MarkKeyAsUsed(scriptAddress []byte) error

	// Fetch the last index for the given key purpose
	// The bool should state whether the key has been used or not
	GetLastKeyIndex(purpose KeyPurpose) (int, bool, error)

	// Returns the first unused path for the given purpose
	GetPathForKey(scriptAddress []byte) (KeyPath, error)

	// Returns an imported private key given a script address
	GetKey(scriptAddress []byte) (*btcec.PrivateKey, error)

	// Returns all imported keys
	GetImported() ([]*btcec.PrivateKey, error)

	// Get a list of unused key indexes for the given purpose
	GetUnused(purpose KeyPurpose) ([]int, error)

	// Fetch all key paths
	GetAll() ([]KeyPath, error)

	// Get the number of unused keys following the last used key
	// for each key purpose.
	GetLookaheadWindows() map[KeyPurpose]int
}

type WatchedScripts interface {
	// Add a script to watch
	Put(scriptPubKey []byte) error

	// Return all watched scripts
	GetAll() ([][]byte, error)

	// Delete a watched script
	Delete(scriptPubKey []byte) error
}

type Utxo struct {
	// Previous txid and output index
	Op wire.OutPoint

	// Block height where this tx was confirmed, 0 for unconfirmed
	AtHeight int32

	// The higher the better
	Value string // ZJ: int64

	// Output script
	ScriptPubkey []byte

	// If true this utxo will not be selected for spending. The primary
	// purpose is track multisig UTXOs which must have separate handling
	// to spend.
	WatchOnly bool
}

func (utxo *Utxo) IsEqual(alt *Utxo) bool {
	if alt == nil {
		return utxo == nil
	}

	if !utxo.Op.Hash.IsEqual(&alt.Op.Hash) {
		return false
	}

	if utxo.Op.Index != alt.Op.Index {
		return false
	}

	if utxo.AtHeight != alt.AtHeight {
		return false
	}

	if utxo.Value != alt.Value {
		return false
	}

	if bytes.Compare(utxo.ScriptPubkey, alt.ScriptPubkey) != 0 {
		return false
	}

	return true
}

type Stxo struct {
	// When it used to be a UTXO
	Utxo Utxo

	// The height at which it met its demise
	SpendHeight int32

	// The tx that consumed it
	SpendTxid chainhash.Hash
}

func (stxo *Stxo) IsEqual(alt *Stxo) bool {
	if alt == nil {
		return stxo == nil
	}

	if !stxo.Utxo.IsEqual(&alt.Utxo) {
		return false
	}

	if stxo.SpendHeight != alt.SpendHeight {
		return false
	}

	if !stxo.SpendTxid.IsEqual(&alt.SpendTxid) {
		return false
	}

	return true
}

type Txn struct {
	Txid 				common.Hash
	Value 				string
	Height 				int32
	Timestamp 			time.Time
	From 				common.Address
	To 					common.Address
	Gas 				string
	Input 				string
	Confirmations 		int64
	Status 				string
}

type KeyPath struct {
	Purpose KeyPurpose
	Index   int
}
