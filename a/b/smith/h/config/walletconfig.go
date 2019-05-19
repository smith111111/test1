package config

import (
	"encoding/json"
	"fmt"
	"os"
)
type WalletConfig struct {
	DataDir    string `json:"data_dir"`
	// is testnet
	Testnet bool    `json:"testnet"`
	// ETH config
	InfuraUrl    string `json:"infura_url"`
	InfuraAPIKey    string `json:"infura_api_key"`
	RegistryAddress  string `json:"registry_address"`
	GalaxyCoinAddress  string `json:"galaxy_coin_address"`
	EtherscanAPI	string	`json:"etherscan_api"`
	EtherscanAPIKey string `json:"etherscan_api_key"`
	// Multi config
	Bitcoin bool    `json:"bitcoin"`
	BitcoinCash bool    `json:"bitcoin_cash"`
	Zcash bool    `json:"zcash"`
	Litecoin bool    `json:"litecoin"`
	// EOS config
	EosClients 	[]string	`json:"eos_apis"`
	EosparkAPI	string	`json:"eos_apis"`
	EosparkAPIKey string `json:"eospark_apikey"`

	SecretBoxWrap       string `json:"secretbox_wrap"`
	SecretBoxCiphertext string `json:"secretbox_ciphertext"`

	// m
	Mnemonic string `json:"-"` // not output
}

func NewWalletConfigFromConfigFile(filename string) (*WalletConfig, error) {
	// 返回一个空config，未存储并且不含助记词。
	config := NewWalletConfig()
	// 打开配置文件
	fl, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	err = json.NewDecoder(fl).Decode(&config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// 从提供的助记词，创建一个新WalletConfig实例。
func NewWalletConfigFromMnemonic(mnemonic string) (*WalletConfig, error) {
	config := NewWalletConfig()
	config.Mnemonic = mnemonic
	return config, nil
}

// 返回一个空vault，未存储并且不含密钥。
func NewWalletConfig() *WalletConfig {
	return &WalletConfig{
		DataDir:".",
		Testnet: true,
		// ETH config
		InfuraUrl: "https://rinkeby.infura.io",		// mainnet "https://mainnet.infura.io"
		InfuraAPIKey: "LDWztu8dxaR6sRgb4rsN",
		RegistryAddress: "0x403d907982474cdd51687b09a8968346159378f3",
		GalaxyCoinAddress: "0x15d3faabb721eff39c0fbfe27c9da8d12974b331",  // mainnet "0x949c97692133b7593e06d7bc5a445dea52665b48"
		EtherscanAPI: "http://api-rinkeby.etherscan.io/api",
		EtherscanAPIKey: "1TPR2KHZKBC5M9MKEN3B38H379XVBHXRM7",
		// Multi config
		Bitcoin: true,
		BitcoinCash: true,
		Zcash: false,
		Litecoin: true,
		Mnemonic:  "",
		// EOS config
		//EosClients: []string{"http://192.168.0.236:8888"},
		EosClients: []string{"https://eosapi.nodepacific.com", "https://api.eosbeijing.one", "https://api.oraclechain.io", "https://eosapi.nodepacific.com", "https://eosbp.atticlab.net"},
		EosparkAPI: "https://api.eospark.com/api",
		EosparkAPIKey: "3f6a08a55e1b096ce114a3a895e1f2ef",
	}
}

// 打印WalletConfig的密钥袋KeyBag中的每个私钥的公钥。
func (c *WalletConfig) PrintMnemonic() {
	fmt.Printf("mnemonic is %s\n", c.Mnemonic)
}

// 将WalletConfig写入到硬盘中去。在写入到文件之前，你必须加密（密封）Mnemonic，否则你可能丢失。
func (c *WalletConfig) WriteToFile(filename string) error {
	// 转换成JSON串字节片段
	cnt, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// 打开文件
	fl, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	// 写入到文件中去
	_, err = fl.Write(cnt)
	if err != nil {
		fl.Close()
		return err
	}

	return fl.Close()
}

// 开箱：解密WalletConfig中的密码箱密文SecretBoxCiphertext，解密的结果反序列化到Mnemonic中去。
func (c *WalletConfig) Open(boxer SecretBoxer) error {
	data, err := boxer.Open(c.SecretBoxCiphertext)
	if err != nil {
		return err
	}
	c.Mnemonic = string(data)

	return nil
}

// 密封：使用密码箱boxer来将WalletConfig中的Mnemonic加密到SecretBoxCiphertext中去。
func (c *WalletConfig) Seal(boxer SecretBoxer) error {
	c.SecretBoxWrap = boxer.WrapType()
	cipherText, err := boxer.Seal([]byte(c.Mnemonic))
	if err != nil {
		return err
	}

	c.SecretBoxCiphertext = cipherText
	return nil
}
