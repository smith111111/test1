package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"time"
)

type Coin interface {
	String() string
	CurrencyCode() string
}

type CoinType uint32

const (
	EOS     CoinType = 0
)

func (c *CoinType) String() string {
	switch *c {
	case EOS:
		return "EOS"
	default:
		return ""
	}
}

func (c *CoinType) CurrencyCode() string {
	switch *c {
	case EOS:
		return "EOS"
	default:
		return ""
	}
}

type Datastore interface {
	Accounts() Accounts
	Txns() Txns
	Keys() Keys
}

type Accounts interface {
	// Put a account to the database
	Put(account Account) error

	Get(key string) (Account, error)

	// Fetch all accounts from the db
	GetAll() ([]Account, error)

	// Delete a account from the db
	Delete(account Account) error
}

type Txns interface {
	// Put a new transaction to the database
	Put(txid string, height int32, timestamp time.Time, value string, symbol string, status string) error

	// Fetch a tx and it's metadata given a hash
	Get(txid chainhash.Hash) (Txn, error)

	// Fetch all transactions from the db
	GetAll() ([]Txn, error)

	// Update the height of a transaction
	UpdateHeight(txid chainhash.Hash, height int32, timestamp time.Time) error

	// Delete a transactions from the db
	Delete(txid *chainhash.Hash) error
}

// Keys provides a database interface for the wallet to save key material, track
// used keys, and manage the look ahead window.
type Keys interface {
	// Import a private key to the database
	ImportKey(key Key) error

	// verify that the private key exists
	HasKey(privateKey string) bool

	// Get the private key with the public key
	GetPrivateKey(publicKey string) (Key, error)

	// Get all imported keys
	GetImported() ([]Key, error)

	// Delete a key from the db
	Delete(key Key) error
}

type Account struct {
	// The Account Name
	Name string

	// The Account Active PublicKey
	ActivePublicKey string

	// The Account Owner PublicKey
	OwnerPublicKey string

	// The Account Authority
	Authority  []string
}

type Txn struct {
	// Transaction ID
	Txid string

	// The height at which it was mined
	Height int32

	// The time the transaction was first seen
	Timestamp time.Time

	// The transaction amount
	Value string

	// The transaction symbol
	Symbol string

	// The transaction status
	Status string

	// The transaction sender
	Sender string

	// The transaction receiver
	Receiver string

	// The transaction memo
	Memo string

	// The transaction confirmations
	Confirmations int32
}

type Key struct {
	// The Account Name
	Name string

	// The Private Key
	PrivateKey string

	// THe Public Key
	PublicKey string
}

type StatusCode string

const (
	StatusUnconfirmed StatusCode = "UNCONFIRMED"
	StatusPending                = "PENDING"
	StatusConfirmed              = "CONFIRMED"
	StatusStuck                  = "STUCK"
	StatusDead                   = "DEAD"
	StatusError                  = "ERROR"
)
