package wallet

import (
	"errors"
	"time"

	"gcwallet/eoswallet/eospark"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	btc "github.com/btcsuite/btcutil"
	hd "github.com/btcsuite/btcutil/hdkeychain"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

// The Wallet interface is used by openbazaar-go for both normal wallet operation (sending
// and receiving) as well as for handling multisig escrow payment as part of its order flow.
// The following is a very high level of the order flow and how it relates to the methods
// described by this interface.

// 1) The buyer clicks a button to place an order with a vendor. In addition to populating
// the order with all the relevant information, the buyer's node calls the `GenerateMultisigScript`
// interface method to generate an address and redeem script that is unique for the order.
// The order is then sent over to the vendor for evaluation.
//
// 2) The vendor receives the order, takes his public key as well as the key provided by the
// buyer and moderator and likewise calls `GenerateMultisigScript` and compares the returned
// address and redeem script to those provide by the buyer in the order to make sure the
// buyer provided valid information. He then sends a message to the buyer notifying that he
// has accepted the order.
//
// 3) The buyer can then either send funds into the multisig address using an external wallet
// or if he wishes to use the built-in wallet, he calls the `Spend` interface method and
// provides the multisig address as the destination.
//
// 4) After the buyer receives the goods he clicks the complete order button in the UI to
// leave a review and release the funds to the vendor. His node calls the `CreateMultisigSignature`
// interface method to generate a signature for the transaction releasing the funds. The signature is
// sent over to the vendor along with his review.
//
// 5) The vendor receives the review and the signature then calls `CreateMultisigSignature`
// himself to generate his signature on the transaction. We now have the two signatures necessary
// to release the funds. The vendor then calls the `Multisign` interface method and includes
// both signatures. The multisign function combines all the signatures into one valid transaction
// then broadcasts it to the network.
//
// The above example is only one possible order flow. There are other variants based on whether or
// not the vendor is online or offline and whether or not the buyer is doing a direct payment or
// escrowed payment.
type Wallet interface {

	// Start is called when the openbazaar-go daemon starts up. At this point in time
	// the wallet implementation should start syncing and/or updating balances, but
	// not before.
	Start()

	// CurrencyCode returns the currency code this wallet implements. For example, "BTC".
	// When running on testnet a `T` should be prepended. For example "TBTC".
	CurrencyCode() string

	// IsDust returns whether the amount passed in is considered dust by network. This
	// method is called when building payout transactions from the multisig to the various
	// participants. If the amount that is supposed to be sent to a given party is below
	// the dust threshold, openbazaar-go will not pay that party to avoid building a transaction
	// that never confirms.
	IsDust(amount int64) bool

	CurrentAccount() *eos.AccountName

	// HasKey returns whether or not the wallet has the key for the given address. This method
	// is called by openbazaar-go when validating payouts from multisigs. It makes sure the
	// transaction that the other party(s) signed does indeed pay to an address that we
	// control.
	HasKey(addr btc.Address) bool

	GetAccount(account eos.AccountName) (*eos.AccountResp, error)

	GetAccounts(publicKey string) (*Accounts, error)

	// Balance returns the confirmed and unconfirmed aggregate balance for the wallet.
	// For utxo based wallets, if a spend of confirmed coins is made, the resulting "change"
	// should be also counted as confirmed even if the spending transaction is unconfirmed.
	// The reason for this that if the spend never confirms, no coins will be lost to the wallet.
	//
	// The returned balances should be in the coin's base unit (for example: satoshis)
	GetBalance(account eos.AccountName) int64

	// Transactions returns a list of transactions for this wallet.
	GetTransactions(symbol, code string, transactionType, sort, page, size int) ([]*Txn, error)

	// GetTransaction return info on a specific transaction given the txid.
	GetTransaction(txid chainhash.Hash) (*Txn, error)

	// ChainTip returns the best block hash and height of the blockchain.
	ChainTip() (uint32, chainhash.Hash)

	GetTokenList() ([]*eospark.Symbol, error)

	// Spend transfers the given amount of coins (in the coin's base unit. For example: in
	// satoshis) to the given address using the provided fee level. Openbazaar-go calls
	// this method in two places. 1) When the user requests a normal transfer from their
	// wallet to another address. 2) When clicking 'pay from internal wallet' to fund
	// an order the user just placed.
	// It also includes a referenceID which basically refers to the order the spend will affect
	Spend(to, memo, amount string) (*chainhash.Hash, error)

	// GenerateMultisigScript should deterministically create a redeem script and address from the information provided.
	// This method should be strictly limited to taking the input data, combining it to produce the redeem script and
	// address and that's it. There is no need to interact with the network or make any transactions when this is called.
	//
	// Openbazaar-go will call this method in the following situations:
	// 1) When the buyer places an order he passes in the relevant keys for each party to get back the address where
	// the funds should be sent and the redeem script. The redeem script is saved in order (and openbazaar-go database).
	//
	// 2) The vendor calls this method when he receives and order so as to validate that the address they buyer is sending
	// funds to is indeed correctly constructed. If this method fails to return the same values for the vendor as it
	// did the buyer, the vendor will reject the order.
	//
	// 3) The moderator calls this function upon receiving a dispute so that he can validate the payment address for the
	// order and make sure neither party is trying to maliciously lie about the details of the dispute to get the moderator
	// to release the funds.
	//
	// Note that according to the order flow, this method is called by the buyer *before* the order is sent to the vendor,
	// and before the vendor validates the order. Only after the buyer hears back from the vendor does the buyer send
	// funds (either from an external wallet or via the `Spend` method) to the address specified in this method's return.
	//
	// `threshold` is the number of keys required to release the funds from the address. If `threshold` is two and len(keys)
	// is three, this is a two of three multisig. If `timeoutKey` is not nil, then the script should allow the funds to
	// be released with a signature from the `timeoutKey` after the `timeout` duration has passed.
	// For example:
	// OP_IF 2 <buyerPubkey> <vendorPubkey> <moderatorPubkey> 3 OP_ELSE <timeout> OP_CHECKSEQUENCEVERIFY <timeoutKey> OP_CHECKSIG OP_ENDIF
	//
	// If `timeoutKey` is nil then the a normal multisig without a timeout should be created.
	GenerateMultisigScript(keys []hd.ExtendedKey, threshold int, timeout time.Duration, timeoutKey *hd.ExtendedKey) (addr btc.Address, redeemScript []byte, err error)

	// SweepAddress should sweep all the funds from the provided inputs into the provided `address` using the given
	// `key`. If `address` is nil, the funds should be swept into an internal address own by this wallet.
	// If the `redeemScript` is not nil, this should be treated as a multisig (p2sh) address and signed accordingly.
	//
	// This method is called by openbazaar-go in the following scenarios:
	// 1) The buyer placed a direct order to a vendor who was offline. The buyer sent funds into a 1 of 2 multisig.
	// Upon returning online the vendor accepts the order and calls SweepAddress to move the funds into his wallet.
	//
	// 2) Same as above but the buyer wishes to cancel the order before the vendor comes online. He calls SweepAddress
	// to return the funds from the 1 of 2 multisig back into has wallet.
	//
	// 3) Same as above but rather than accepting the order, the vendor rejects it. When the buyer receives the reject
	// message he calls SweepAddress to move the funds back into his wallet.
	//
	// 4) The timeout has expired on a 2 of 3 multisig. The vendor calls SweepAddress to claim the funds.
	//SweepAddress(ins []TransactionInput, address *btc.Address, key *hd.ExtendedKey, redeemScript *[]byte, feeLevel FeeLevel) (*chainhash.Hash, error)

	// CreateMultisigSignature should build a transaction using the given inputs and outputs and sign it with the
	// provided key. A list of signatures (one for each input) should be returned.
	//
	// This method is called by openbazaar-go by each party whenever they decide to release the funds from escrow.
	// This method should not actually move any funds or make any transactions, only create necessary signatures to
	// do so. The caller will then take the signature and share it with the other parties. Once all parties have shared
	// their signatures, the person who wants to release the funds collects them and uses them as an input to the
	// `Multisign` method.
	//CreateMultisigSignature(ins []TransactionInput, outs []TransactionOutput, key *hd.ExtendedKey, redeemScript []byte, feePerByte uint64) ([]Signature, error)

	// Multisign collects all of the signatures generated by the `CreateMultisigSignature` function and builds a final
	// transaction that can then be broadcast to the blockchain. The []byte return is the raw transaction. It should be
	// broadcasted if `broadcast` is true. If the signatures combine and produce an invalid transaction then an error
	// should be returned.
	//
	// This method is called by openbazaar-go by whichever party to the escrow is trying to release the funds only after
	// all needed parties have signed using `CreateMultisigSignature` and have shared their signatures with each other.
	//Multisign(ins []TransactionInput, outs []TransactionOutput, sigs1 []Signature, sigs2 []Signature, redeemScript []byte, feePerByte uint64, broadcast bool) ([]byte, error)

	// AddWatchedAddress adds an address to the wallet to get notifications back when coins
	// are received or spent from it. These watch only addresses should be persisted between
	// sessions and upon each startup the wallet should be made to listen for transactions
	// involving them.
	AddWatchedAddress(addr btc.Address) error

	// AddTransactionListener is how openbazaar-go registers to receive a callback whenever
	// a transaction is received that is relevant to this wallet or any of its watch only
	// addresses. An address is considered relevant if any inputs or outputs match an address
	// owned by this wallet, or being watched by the wallet via AddWatchedAddress method.
	AddTransactionListener(callback func(TransactionCallback))

	// ReSyncBlockchain is called in response to a user action to rescan transactions. API based
	// wallets should do another scan of their addresses to find anything missing. Full node, or SPV
	// wallets should rescan/re-download blocks starting at the fromTime.
	ReSyncBlockchain(fromTime time.Time)

	// GetConfirmations returns the number of confirmations and the height for a transaction.
	GetConfirmations(txid chainhash.Hash) (confirms, atHeight uint32, err error)

	// ExchangeRates returns an ExchangeRates implementation which will provide
	// fiat exchange rate data for this coin.
	ExchangeRates() ExchangeRates

	// Close should cleanly disconnect from the wallet and finish writing
	// anything it needs to to disk.
	Close()
}

// The end leaves on the HD wallet have only two possible values. External keys are those given
// to other people for the purpose of receiving transactions. These may include keys used for
// refund addresses. Internal keys are used only by the wallet, primarily for change addresses
// but could also be used for shuffling around UTXOs.
type KeyPurpose int

const (
	EXTERNAL KeyPurpose = 0
	INTERNAL            = 1
)

type AccountsResp struct {
	AccountNames    []eos.AccountName 	`json:"account_names"`
}

type ImportedKey struct {
	AccountName 	eos.AccountName 	`json:"account_name"`
	ActivePublicKey ecc.PublicKey		`json:"active_public_key"`
	OwnerPublicKey 	ecc.PublicKey		`json:"owner_public_key"`
	Authority   	[]string			`json:"authority"`
}

// This callback is passed to any registered transaction listeners when a transaction is detected
// for the wallet.
type TransactionCallback struct {
	Txid      string
	Outputs   []TransactionOutput
	Inputs    []TransactionInput
	Height    int32
	Timestamp time.Time
	Value     int64
	WatchOnly bool
	BlockTime time.Time
}

type TransactionOutput struct {
	Address btc.Address
	Value   int64
	Index   uint32
	OrderID string
}

type TransactionInput struct {
	OutpointHash  []byte
	OutpointIndex uint32
	LinkedAddress btc.Address
	Value         int64
	OrderID       string
}

// OpenBazaar uses p2sh addresses for escrow. This object can be used to store a record of a
// transaction going into or out of such an address. Incoming transactions should have a positive
// value and be market as spent when the UXTO is spent. Outgoing transactions should have a
// negative value. The spent field isn't relevant for outgoing transactions.
type TransactionRecord struct {
	Txid      string
	Index     uint32
	Value     int64
	Address   string
	Spent     bool
	Timestamp time.Time
}

// This object contains a single signature for a multisig transaction. InputIndex specifies
// the index for which this signature applies.
type Signature struct {
	InputIndex uint32
	Signature  []byte
}

// Errors
var (
	ErrorInsuffientFunds error = errors.New("Insuffient funds")
	ErrorDustAmount      error = errors.New("Amount is below network dust treshold")
)
