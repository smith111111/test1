package gcwallet

import (
	"fmt"
	bwi "gcwallet/btc-wallet-interface"
	wc "gcwallet/config"
	ethwi "gcwallet/eth-wallet-interface"
	ethdb "gcwallet/go-ethwallet/boltdb"
	ethconfig "gcwallet/go-ethwallet/config"
	eth "gcwallet/go-ethwallet/wallet"
	"gcwallet/multiwallet"
	"gcwallet/multiwallet/api"
	multiconfig "gcwallet/multiwallet/config"
	eoswi "gcwallet/eos-wallet-interface"
	eosdb "gcwallet/eoswallet/boltdb"
	eosconfig "gcwallet/eoswallet/config"
	eos "gcwallet/eoswallet/wallet"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"
	"os"
	"path"
	"strings"
	"time"
	"strconv"
	"unicode"
	"github.com/tyler-smith/go-bip39/wordlists"
	"gcwallet/log"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(data []byte) (err error) {
	unix, err := strconv.ParseInt(string(data), 10, 64)
	tm := time.Unix(unix, 0)
	*t = Timestamp(tm)
	return
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	unix := strconv.FormatInt(time.Time(t).Unix(),10)
	return []byte(unix), nil
}

func (t Timestamp) String() string {
	return strconv.FormatInt(time.Time(t).Unix(),10)
}

// ethAddress implements the WalletAddress interface
type ethAddress struct {
	address *common.Address
}

// String representation of eth address
func (addr ethAddress) String() string {
	return addr.address.String()
}

// EncodeAddress returns hex representation of the address
func (addr ethAddress) EncodeAddress() string {
	return addr.address.Hex()
}

// ScriptAddress returns byte representation of address
func (addr ethAddress) ScriptAddress() []byte {
	return addr.address.Bytes()
}

// IsForNet returns true because EthAddress has to become btc.Address
func (addr ethAddress) IsForNet(params *chaincfg.Params) bool {
	return true
}

type GalaxyCoinWallet struct {
	ethWallet   *eth.EthereumWallet
	gctWallet    *eth.ERC20Wallet
	multiWallet *multiwallet.MultiWallet
	eosWallet 	*eos.EosWallet
	scatter     *Scatter
}

func InitLogPath(path string) {
	log.SetLogPath(path)
}

func New() *GalaxyCoinWallet {
	gcw := &GalaxyCoinWallet{}
	//
	//rinkebyUrl := fmt.Sprintf("https://rinkeby.infura.io/%s", eth.InfuraAPIKey)
	//mnemonic := "label pyramid flat spike course crystal humor throw rug frozen food comic" //"soup arch join universe table nasty fiber solve hotel luggage double clean tell oppose hurry weather isolate decline quick dune song enforce curious menu" // "wolf dragon lion stage rose snow sand snake kingdom hand daring flower foot walk sword"
	//
	//// 设置ETH钱包
	//ethCfg := config.CoinConfig{}
	//ethCfg.ClientAPIs = []string{rinkebyUrl}
	//ethCfg.CoinType = wi.Ethereum
	//ethCfg.Options = make(map[string]interface{})
	//ethCfg.Options["RegistryAddress"] = "0x403d907982474cdd51687b09a8968346159378f3" //"0xab8dd0e05b73529b440d9c9df00b5f490c8596ff"
	//gcw.ethWallet, _ = eth.NewEthereumWallet(ethCfg, mnemonic, nil)
	//
	//// 设置GalaxyCoin钱包
	//gcCfg := config.CoinConfig{}
	//gcCfg.ClientAPIs = []string{rinkebyUrl}
	//gcCfg.CoinType = wi.Ethereum
	//gcCfg.Options = make(map[string]interface{})
	//gcCfg.Options["RegistryAddress"] = "0x403d907982474cdd51687b09a8968346159378f3"
	//gcCfg.Options["Name"] = "GalaxyCoin"
	//gcCfg.Options["Symbol"] = "GC"
	//gcCfg.Options["MainNetAddress"] = "0x15d3faabb721eff39c0fbfe27c9da8d12974b331" // mainnet "0x949c97692133b7593e06d7bc5a445dea52665b48"
	//gcCfg.Options["RinkebyAddress"] = "0x15d3faabb721eff39c0fbfe27c9da8d12974b331"
	//gcw.galaxyCoinWallet, _ = eth.NewERC20Wallet(gcCfg, mnemonic, nil)
	//
	//// 设置多币种钱包
	//m := make(map[wi.CoinType]bool)
	//m[wi.Bitcoin] = true
	//m[wi.BitcoinCash] = true
	//m[wi.Zcash] = true
	//m[wi.Litecoin] = true
	////params := &chaincfg.MainNetParams
	//params := &chaincfg.TestNet3Params
	//defCfg := config.NewDefaultConfig(m, params)
	//defCfg.Mnemonic = mnemonic
	////var err error
	//mw, _ := multiwallet.NewMultiWallet(defCfg)
	////if err != nil {
	////	return err
	////}
	//gcw.multiWallet = &mw

	gcw.scatter = &Scatter{}

	return gcw
}

func mustConfigDir(configdir string) error {
	dir := path.Join(configdir, ".gcwallet")
	err := os.Mkdir(dir, 0700)
	// .gcwallet directory already exist?
	if os.IsExist(err) {
		err = nil // then nullify the error
	}
	if err != nil {
		return err
	}

	fmt.Println(".gcwallet directory created")

	return nil
}

func removeConfigDir(configdir string) error {
	dir := path.Join(configdir, ".gcwallet")
	err := os.RemoveAll(dir)
	if err == nil || os.IsNotExist(err) {
		fmt.Println(".gcwallet directory already removed/deleted")
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Println(".gcwallet directory removed/deleted")

	return nil
}

//func (gcw *GalaxyCoinWallet) writeConfig(filename string, force bool) error {
//	jww.INFO.Println("Attempting to write configuration to file.")
//	ext := filepath.Ext(filename)
//	if len(ext) <= 1 {
//		return fmt.Errorf("Filename: %s requires valid extension.", filename)
//	}
//	configType := ext[1:]
//	if !stringInSlice(configType, SupportedExts) {
//		return UnsupportedConfigError(configType)
//	}
//	if v.config == nil {
//		v.config = make(map[string]interface{})
//	}
//	var flags int
//	if force == true {
//		flags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY
//	} else {
//		if _, err := os.Stat(filename); os.IsNotExist(err) {
//			flags = os.O_WRONLY
//		} else {
//			return fmt.Errorf("File: %s exists. Use WriteConfig to overwrite.", filename)
//		}
//	}
//	f, err := v.fs.OpenFile(filename, flags, os.FileMode(0644))
//	if err != nil {
//		return err
//	}
//	return v.marshalWriter(f, configType)
//}

func (gcw *GalaxyCoinWallet) StartScatter() bool {
	log.WriteLog("GalaxyCoinWallet, StartScatter...")
	log.WriteLog("GalaxyCoinWallet, ==================================")
	return gcw.scatter.Start()
}

func (gcw *GalaxyCoinWallet) StopScatter() bool {
	log.WriteLog("GalaxyCoinWallet, StopScatter...")
	return gcw.scatter.Stop()
}

func (gcw *GalaxyCoinWallet) HasConfig(configdir string) bool {
	exists := false
	dir := path.Join(configdir, ".gcwallet")
	if _, err := os.Stat(dir); err == nil {
		exists = true
	}
	return exists
}

// 加载钱包
func (gcw *GalaxyCoinWallet) LoadConfig(configdir, password string) string {
	// 配置文件包文件名
	configFile := path.Join(configdir, ".gcwallet", "gcwallet.conf")
	// 从提供的配置文件返回一个新的WalletConfig实例。
	walletConfig, err := wc.NewWalletConfigFromConfigFile(configFile)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	walletConfig.DataDir = configdir

	// 返回一个指定类型的秘密装箱器
	boxer, err := wc.SecretBoxerForType(walletConfig.SecretBoxWrap, password)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	// unlock
	err = walletConfig.Open(boxer)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	//// 打印Mnemonic。
	//walletConfig.PrintMnemonic()

	// lock
	err = walletConfig.Seal(boxer)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	// 将配置写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失Mnemonic。
	err = walletConfig.WriteToFile(configFile)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	// 加载钱包
	rinkebyUrl := fmt.Sprintf("%s/%s", walletConfig.InfuraUrl, walletConfig.InfuraAPIKey)
	//mnemonic := "label pyramid flat spike course crystal humor throw rug frozen food comic" //"soup arch join universe table nasty fiber solve hotel luggage double clean tell oppose hurry weather isolate decline quick dune song enforce curious menu" // "wolf dragon lion stage rose snow sand snake kingdom hand daring flower foot walk sword"

	// 设置ETH钱包
	ethCfg := ethconfig.CoinConfig{}
	ethCfg.ClientAPIs = []string{rinkebyUrl}
	ethCfg.CoinType = ethwi.Ethereum
	ethCfg.Options = make(map[string]interface{})
	ethCfg.Options["RegistryAddress"] = walletConfig.RegistryAddress
	ethCfg.Options["EtherscanAPI"] = walletConfig.EtherscanAPI
	ethCfg.Options["EtherscanAPIKey"] = walletConfig.EtherscanAPIKey
	var ethereumDS *ethdb.BoltDatastore
	if walletConfig.Testnet {
		ethereumDS, _ = ethdb.Create(walletConfig.DataDir,"rinkeby")
	} else {
		ethereumDS, _ = ethdb.Create(walletConfig.DataDir, "ethereum")
	}
	ethCfg.DB = ethereumDS
	gcw.ethWallet, err = eth.NewEthereumWallet(ethCfg, walletConfig.Mnemonic, nil)
	if err != nil {
		return fmt.Sprintf("error: eth wallet %s", err.Error())
	}

	// 设置GalaxyCoin钱包
	gcCfg := ethconfig.CoinConfig{}
	gcCfg.ClientAPIs = []string{rinkebyUrl}
	gcCfg.CoinType = ethwi.Ethereum
	gcCfg.Options = make(map[string]interface{})
	gcCfg.Options["RegistryAddress"] = walletConfig.RegistryAddress
	gcCfg.Options["EtherscanAPI"] = walletConfig.EtherscanAPI
	gcCfg.Options["EtherscanAPIKey"] = walletConfig.EtherscanAPIKey
	gcCfg.Options["Name"] = "GalaxyCoin"
	gcCfg.Options["Symbol"] = "GC"
	gcCfg.Options["MainNetAddress"] = walletConfig.GalaxyCoinAddress // mainnet "0x949c97692133b7593e06d7bc5a445dea52665b48"
	gcCfg.Options["RinkebyAddress"] = walletConfig.GalaxyCoinAddress
	var gcDS *ethdb.BoltDatastore
	if walletConfig.Testnet {
		gcDS, _ = ethdb.Create(walletConfig.DataDir, "gcrinkeby")
	} else {
		gcDS, _ = ethdb.Create(walletConfig.DataDir, "gcethereum")
	}
	gcCfg.DB = gcDS
	gcw.gctWallet, err = eth.NewERC20Wallet(gcCfg, walletConfig.Mnemonic, nil)
	if err != nil {
		return fmt.Sprintf("error: gc wallet %s", err.Error())
	}

	// 设置多币种钱包
	m := make(map[bwi.CoinType]bool)
	m[bwi.Bitcoin] = walletConfig.Bitcoin
	m[bwi.BitcoinCash] = walletConfig.BitcoinCash
	m[bwi.Zcash] = walletConfig.Zcash
	m[bwi.Litecoin] = walletConfig.Litecoin
	params := &chaincfg.MainNetParams
	if walletConfig.Testnet {
		params = &chaincfg.TestNet3Params
	}
	defCfg := multiconfig.NewDefaultConfig(walletConfig.DataDir, m, params)
	defCfg.Mnemonic = walletConfig.Mnemonic
	mw, err := multiwallet.NewMultiWallet(defCfg)
	if err != nil {
		return fmt.Sprintf("error: btc wallet %s", err.Error())
	}
	gcw.multiWallet = &mw

	// 设置EOS钱包
	eosCfg := eosconfig.CoinConfig{}
	eosCfg.ClientAPIs = walletConfig.EosClients
	eosCfg.CoinType = eoswi.EOS
	eosCfg.Options = make(map[string]interface{})
	eosCfg.Options["EosparkAPI"] = walletConfig.EosparkAPI
	eosCfg.Options["EosparkAPIKey"] = walletConfig.EosparkAPIKey
	var eosDS *eosdb.BoltDatastore
	if walletConfig.Testnet {
		eosDS, _ = eosdb.Create(walletConfig.DataDir, "eosx")
	} else {
		eosDS, _ = eosdb.Create(walletConfig.DataDir, "eos")
	}
	eosCfg.DB = eosDS
	gcw.eosWallet, err = eos.NewEosWallet(eosCfg, configdir, password, true)
	if err != nil {
		return fmt.Sprintf("error: eos wallet %s", err.Error())
	}

	return "load successfully"
}

func (gcw *GalaxyCoinWallet) IsMnemonicValid(mnemonic string) bool {
	// 如果是中文助记词需要手动设置单词列表为中文
	if gcw.IsHanMnemonic(mnemonic) {
		bip39.SetWordList(wordlists.ChineseSimplified)
	} else {
		bip39.SetWordList(wordlists.English)
	}
	if !bip39.IsMnemonicValid(mnemonic) {
		return false
	}
	return true
}

func (gcw *GalaxyCoinWallet) IsHanMnemonic(mnemonic string) bool {
	if mnemonic == "" {
		return false
	}

	// 将空格都去除，判断是否为中文助记词
	mnemonic = strings.Replace(mnemonic, " ", "", -1)
	for _, str := range mnemonic {
		if !unicode.Is(unicode.Han, str) {
			return false
		}
	}
	return true
}

func (gcw *GalaxyCoinWallet) CreateConfig(configdir, mnemonic, password string) string {
	if !gcw.IsMnemonicValid(mnemonic){
		return fmt.Sprintf("error: invalid mnemonic %s.", mnemonic)
	}

	// 配置文件包文件名
	configFile := path.Join(configdir, ".gcwallet", "gcwallet.conf")
	// 如果钱包文件已经存在了，则退出
	if _, err := os.Stat(configFile); err == nil {
		// 从提供的配置文件返回一个新的WalletConfig实例。
		walletConfig, err := wc.NewWalletConfigFromConfigFile(configFile)
		if err != nil {
			return fmt.Sprintf("error: %s", err.Error())
		}
		walletConfig.DataDir = configdir

		// 返回一个指定类型的秘密装箱器
		boxer, err := wc.SecretBoxerForType(walletConfig.SecretBoxWrap, password)
		if err != nil {
			return fmt.Sprintf("error: %s", err.Error())
		}

		// unlock
		err = walletConfig.Open(boxer)
		if err != nil {
			return fmt.Sprintf("error: %s", err.Error())
		}
		if strings.ToLower(mnemonic) == strings.ToLower(walletConfig.Mnemonic) {
			return fmt.Sprintf("Wallet file %s already exists, rename it before running `create`.", configFile)
		}
	}

	// 删除旧数据
	err := removeConfigDir(configdir)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	err = mustConfigDir(configdir)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	// 返回一个空walletconfig，未存储Mnemonic。
	walletConfig := wc.NewWalletConfig()
	walletConfig.Mnemonic = mnemonic
	walletConfig.DataDir = configdir

	// 新建一个密码装箱器实例
	var boxer wc.SecretBoxer
	boxer = wc.NewPassphraseBoxer(password)

	// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
	err = walletConfig.Seal(boxer)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	// 将配置写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
	err = walletConfig.WriteToFile(configFile)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return "create successfully"
}

func (gcw *GalaxyCoinWallet) Start() string {
	gcw.ethWallet.Start()
	gcw.gctWallet.Start()

	// 启动比特币钱包
	go api.ServeAPI(*gcw.multiWallet)
	gcw.multiWallet.Start()

	gcw.eosWallet.Start()

	return "success"
}

func (gcw *GalaxyCoinWallet) Close() string {
	gcw.ethWallet.Close()
	gcw.gctWallet.Close()
	gcw.multiWallet.Close()
	gcw.eosWallet.Close()

	return "success"
}

func (gcw *GalaxyCoinWallet) NewMnemonic() string {
	ent, err := bip39.NewEntropy(128)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	mnemonic, err := bip39.NewMnemonic(ent)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return mnemonic
}
