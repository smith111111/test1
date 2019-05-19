package api

type IdentityRsp struct {
	Hash      string             `json:"hash"`
	PublicKey string             `json:"publicKey"`
	Name      string             `json:"name"`
	Kyc       bool               `json:"kyc"`
	Accounts  []*IdentityAccount `json:"accounts"`
}

type IdentityAccount struct {
	Name       string `json:"name"`
	Authority  string `json:"authority"`
	PublicKey  string `json:"publicKey"`
	Blockchain string `json:"blockchain"`
	ChainId    string `json:"chainId"`
	IsHardware bool   `json:"isHardware"`
}

type Payload struct {
	Origin string `json:"origin"`
	Fields []byte `json:"fields"`
}
