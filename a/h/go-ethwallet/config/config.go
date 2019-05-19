package config

import (
	"time"

	"gcwallet/eth-wallet-interface"
	"gcwallet/go-ethwallet/cache"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/op/go-logging"
	"golang.org/x/net/proxy"
)

type Config struct {
	// Network parameters. Set mainnet, testnet, or regtest using this.
	Params *chaincfg.Params

	// Bip39 mnemonic string. If empty a new mnemonic will be created.
	Mnemonic string

	// The date the wallet was created.
	// If before the earliest checkpoint the chain will be synced using the earliest checkpoint.
	CreationDate time.Time

	// A Tor proxy can be set here causing the wallet will use Tor
	Proxy proxy.Dialer

	// A logger. You can write the logs to file or stdout or however else you want.
	Logger logging.Backend

	// Cache is a persistable storage provided by the consumer where the wallet can
	// keep state between runtime executions
	Cache cache.Cacher

	// A list of coin configs. One config should be included for each coin to be used.
	Coins []CoinConfig

	// Disable the exchange rate functionality in each wallet
	DisableExchangeRates bool
}

type CoinConfig struct {
	// The type of coin to configure
	CoinType wallet.CoinType

	// The default fee-per-byte for each level
	LowFee    uint64
	MediumFee uint64
	HighFee   uint64

	// The highest allowable fee-per-byte
	MaxFee uint64

	// External API to query to look up fees. If this field is nil then the default fees will be used.
	// If the API is unreachable then the default fees will likewise be used. If the API returns a fee
	// greater than MaxFee then the MaxFee will be used in place. The API response must be formatted as
	// { "fastestFee": 40, "halfHourFee": 20, "hourFee": 10 }
	FeeAPI string

	// The trusted APIs to use for querying for balances and listening to blockchain events.
	ClientAPIs []string

	// An implementation of the Datastore interface for each desired coin
	DB wallet.Datastore

	// Custom options for wallet to use
	Options map[string]interface{}
}