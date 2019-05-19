package vault

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	saltLength         = 16
	nonceLength        = 24
	keyLength          = 32
	shamirSecretLength = 32
)

func deriveKey(passphrase string, salt []byte) [keyLength]byte {
	// argon2被用来从一个密码来派生加密密钥
	secretKeyBytes := argon2.IDKey([]byte(passphrase), salt, 4, 64*1024, 4, 32)
	var secretKey [keyLength]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}

// 密码装箱器
type PassphraseBoxer struct {
	// 使用一个密码来进行加密、解密工作。
	passphrase string
}

func NewPassphraseBoxer(password string) *PassphraseBoxer {
	return &PassphraseBoxer{
		passphrase: password,
	}
}

func (b *PassphraseBoxer) WrapType() string {
	// 装箱类型为密码串
	return "passphrase"
}

// 加密：将字节片段in进行加密并返回一个base64字符串
func (b *PassphraseBoxer) Seal(in []byte) (string, error) {
	var nonce [nonceLength]byte
	// 生成一个随机nonce
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	salt := make([]byte, saltLength)
	// 生成一个随机salt
	if _, err := crypto_rand.Read(salt); err != nil {
		return "", err
	}

	// 根据密码、盐，派生一个密钥
	secretKey := deriveKey(b.passphrase, salt)
	// 前缀为：salt + nonce
	prefix := append(salt, nonce[:]...)

	// 密文：加密的身份验证的in的拷贝被追加打prefix尾部，并返回。
	cipherText := secretbox.Seal(prefix, in, &nonce, &secretKey)

	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

// 解密：将加密的base64串in解密，返回解密的字节片段。
func (b *PassphraseBoxer) Open(in string) ([]byte, error) {
	buf, err := base64.RawStdEncoding.DecodeString(in)
	if err != nil {
		return []byte{}, err
	}

	salt := make([]byte, saltLength)
	// 从密文中取出salt
	copy(salt, buf[:saltLength])
	var nonce [nonceLength]byte
	// 从密文中取出nonce
	copy(nonce[:], buf[saltLength:nonceLength+saltLength])

	// 根据密码、盐，派生一个密钥
	secretKey := deriveKey(b.passphrase, salt)
	// 解密剩余部分
	decrypted, ok := secretbox.Open(nil, buf[nonceLength+saltLength:], &nonce, &secretKey)
	if !ok {
		return []byte{}, fmt.Errorf("failed to decrypt")
	}
	return decrypted, nil
}
