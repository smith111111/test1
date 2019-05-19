package boltdb

type Key struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	ScriptAddress string `storm:"unique" json:"script_address"`
	Purpose int `json:"purpose"`
	KeyIndex int `json:"key_index"`
	Used bool `json:"used"`
	Key string `json:"key"`
}

type Utxo struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Outpoint string `storm:"unique" json:"outpoint"`
	Value int64 `json:"value"`
	Height int `json:"height"`
	ScriptPubkey string `json:"script_pubkey"`
	WatchOnly bool `json:"watch_only"`
}


type Stxo struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Outpoint string `storm:"unique" json:"outpoint"`
	Value int64 `json:"value"`
	Height int `json:"height"`
	ScriptPubkey string `json:"script_pubkey"`
	WatchOnly bool `json:"watch_only"`
	SpendHeight int `json:"spend_height"`
	SpendTxid string `json:"spend_txid"`
}

type Txn struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Txid string `storm:"unique" json:"txid"`
	Value int64 `json:"value"`
	Height int `json:"height"`
	Timestamp int `json:"timestamp"`
	WatchOnly bool `json:"watch_only"`
	Tx []byte `json:"tx"`
}

type WatchedScript struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	ScriptPubkey string `storm:"unique" json:"script_pubkey"`
}

type Config struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Key string `json:"key"`
	Value string `json:"value"`
}