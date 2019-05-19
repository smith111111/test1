package boltdb

type Key struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Name string `json:"name"`
	PermName string	`json:"perm_name"`
	PrivateKey string `json:"private_key"`
	PublicKey string `json:"public_key"`
}

type Account struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Name string `json:"name"`
	ActivePublicKey string `json:"public_key"json:"active_public_key"`
	OwnerPublicKey string `json:"owner_public_key"`
	Authority  string `json:"authority"`
}

type Txn struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Txid string `storm:"unique" json:"txid"`
	Height int32 `json:"height"`
	Timestamp int64 `json:"timestamp"`
	Value string `json:"value"`
	Symbol string `json:"symbol"`
	Status string `json:"status"`
}

type Config struct {
	ID int `storm:"id,increment"` // primary key with auto increment
	Key string `json:"key"`
	Value string `json:"value"`
}