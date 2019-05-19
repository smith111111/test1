package config

import (
	"fmt"
)

// 秘密装箱器接口
type SecretBoxer interface {
	// 密封：加密in
	Seal(in []byte) (string, error)
	// 开箱：解密in
	Open(in string) ([]byte, error)
	// 装箱器类型：例如，密码。
	WrapType() string
}

// 返回一个指定类型的秘密装箱器
func SecretBoxerForType(boxerType string, password string) (SecretBoxer, error) {
	switch boxerType {
		case "passphrase":
			return NewPassphraseBoxer(password), nil
		default:
			return nil, fmt.Errorf("unknown secret boxer: %s", boxerType)
	}
}
