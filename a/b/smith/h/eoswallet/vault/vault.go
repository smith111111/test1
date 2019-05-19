package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"os"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"path"
)

// Vault 代表一个`eosgc`钱包。它包含加密的资料，以便加载一个KeyBag，
// 那是使用`eos-go` 库（包括嵌套的keosd兼容的钱包）为签名交易的签名提供者。
type Vault struct {
	// 类型
	Kind    string `json:"kind"`
	// 版本
	Version int    `json:"version"`
	// 备注
	Comment string `json:"comment"`

	// 密码装箱器类型
	SecretBoxWrap       string `json:"secretbox_wrap"`
	// 秘密装箱器密文
	SecretBoxCiphertext string `json:"secretbox_ciphertext"`

	// 密钥袋
	KeyBag *eos.KeyBag `json:"-"` // 不输出到JSON
}

// 从提供的eos钱包文件名返回一个新的Vault实例。
func NewVaultFromWalletFile(filename string) (*Vault, error) {
	// 返回一个空vault，未存储并且不含密钥。
	v := NewVault()
	// 打开eos钱包文件
	fl, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	err = json.NewDecoder(fl).Decode(&v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// 从提供的keys文件名（格式：每行一个单个私钥），创建一个新Vault实例。
func NewVaultFromKeysFile(keysFile string) (*Vault, error) {
	v := NewVault()
	if err := v.KeyBag.ImportFromFile(keysFile); err != nil {
		return nil, err
	}
	return v, nil
}

// 从提供的单个私钥串，创建一个新Vault实例。
func NewVaultFromSingleKey(privKey string) (*Vault, error) {
	v := NewVault()
	key, err := ecc.NewPrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("import private key: %s", err)
	}
	v.KeyBag.Keys = append(v.KeyBag.Keys, key)
	return v, nil
}

// 从提供的助记词串，创建一个新Vault实例。
func NewVaultFromMnemonic(mnemonic string, password string) (*Vault, error) {
	v := NewVault()
	seed := bip39.NewSeed(mnemonic, password)
	key, err:= ecc.NewDeterministicPrivateKey(bytes.NewReader(seed))
	if err != nil {
		return nil, fmt.Errorf("import private key: %s", err)
	}
	v.KeyBag.Keys = append(v.KeyBag.Keys, key)
	return v, nil
}

// 返回一个空vault，未存储并且不含密钥。
func NewVault() *Vault {
	return &Vault{
		Kind:    "eos-wallet",
		Version: 1,
		KeyBag:  eos.NewKeyBag(),
	}
}

// 创建一个新的EOS密钥对，保存私钥在本地钱包中，并返回公钥。它不存储这个钱包，你最好是马上做保存。
func (v *Vault) NewKeyPair() (pub ecc.PublicKey, err error) {
	privKey, err := ecc.NewRandomPrivateKey()
	if err != nil {
		return
	}

	v.KeyBag.Keys = append(v.KeyBag.Keys, privKey)

	pub = privKey.PublicKey()
	return
}

// 追加提供的私钥到Vault的密钥袋KeyBag中去，并返回公钥。
func (v *Vault) AddPrivateKey(privateKey *ecc.PrivateKey) (pub ecc.PublicKey) {
	v.KeyBag.Keys = append(v.KeyBag.Keys, privateKey)
	pub = privateKey.PublicKey()
	return
}

// 打印Vault的密钥袋KeyBag中的每个私钥的公钥。
func (v *Vault) PrintPublicKeys() {
	fmt.Printf("Public keys contained within (%d in total):\n", len(v.KeyBag.Keys))
	for _, key := range v.KeyBag.Keys {
		fmt.Println("-", key.PublicKey().String())
	}
}

// 打印Vault的密钥袋KeyBag中的每个私钥 - 公钥。
func (v *Vault) PrintPrivateKeys() {
	fmt.Printf("Private keys contained within (%d in total):\n", len(v.KeyBag.Keys))
	for _, key := range v.KeyBag.Keys {
		fmt.Printf("- %s (corresponds to %s)\n", key, key.PublicKey())
	}
}

func (v *Vault) AddPrivateKeyAndWriteToFile(privateKey *ecc.PrivateKey, configDir, password string) (*ecc.PublicKey, error) {
	// 配置文件包文件名
	walletFile := path.Join(configDir, ".gcwallet", "gcwallet-eos")

	// 获取秘密装箱器
	boxer, err := SecretBoxerForType(v.SecretBoxWrap, password)
	if err != nil {
		return nil, fmt.Errorf("add private key secret boxer error: %s", err.Error())
	}

	//// 解锁
	//err = v.Open(boxer)
	//if err != nil {
	//	return nil, fmt.Errorf("add private key open vault error:", err)
	//}

	publicKey := v.AddPrivateKey(privateKey)

	// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
	if err := v.Seal(boxer); err != nil {
		return nil, fmt.Errorf("add private key seal vault error: %s, password is %s", err.Error(), password)
	}

	// 将Vault写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
	if err := v.WriteToFile(walletFile); err != nil {
		return nil, fmt.Errorf("add private key write to file error: %s", err.Error())
	}

	return &publicKey, nil

}

// 将Vault写入到硬盘中去。在写入到文件之前，你必须加密（密封），否则你可能丢失。
func (v *Vault) WriteToFile(filename string) error {
	// 转换成JSON串字节片段
	cnt, err := json.MarshalIndent(v, "", "  ")
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

// 开箱：解密Vault中的密码箱密文SecretBoxCiphertext，解密的结果反序列化到KeyBag中去。
func (v *Vault) Open(boxer SecretBoxer) error {
	data, err := boxer.Open(v.SecretBoxCiphertext)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, v.KeyBag)
	if err != nil {
		return err
	}

	return nil
}

// 密封：使用密码箱boxer来将Vault中的KeyBag加密到SecretBoxCiphertext中去。
func (v *Vault) Seal(boxer SecretBoxer) error {
	payload, err := json.Marshal(v.KeyBag)
	if err != nil {
		return err
	}

	v.SecretBoxWrap = boxer.WrapType()
	cipherText, err := boxer.Seal(payload)
	if err != nil {
		return err
	}

	v.SecretBoxCiphertext = cipherText
	return nil
}
