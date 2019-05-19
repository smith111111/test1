package wallet

import (
	crypto_rand "crypto/rand"
	"fmt"
	"math/big"
	"os"
	"errors"
	"golang.org/x/crypto/ssh/terminal"
)

func GetPassword(prompt string) (string, error) {
	fd := os.Stdin.Fd()
	fmt.Printf(prompt)
	pass, err := terminal.ReadPassword(int(fd))
	fmt.Println("")
	return string(pass), err
}

// GetConfirmation will prompt for a 4 random number from
// 1000-9999. The `prompt` should always contain a %d to display the
// confirmation security code.
func GetConfirmation(prompt string) (bool, error) {
	value, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(8999))
	if err != nil {
		return false, err
	}

	//随机验证码
	randVal := 1000 + value.Int64()

	// 提示输入验证码
	pw, err := GetPassword(fmt.Sprintf(prompt, randVal))
	if err != nil {
		return false, err
	}

	confirmPW := fmt.Sprintf("%d", randVal)

	return pw == confirmPW, nil
}


func GetDecryptPassphrase() (string, error) {
	passphrase, err := GetPassword("Enter passphrase to decrypt your vault: ")
	if err != nil {
		return "", fmt.Errorf("reading password: %s", err)
	}

	return passphrase, nil
}
func GetEncryptPassphrase() (string, error) {
	// 提示输入一个密码
	passphrase, err := GetPassword("Enter passphrase to encrypt your vault: ")
	if err != nil {
		return "", fmt.Errorf("reading password: %s", err)
	}

	// 提示输入密码确认
	passphraseConfirm, err := GetPassword("Confirm passphrase: ")
	if err != nil {
		return "", fmt.Errorf("reading confirmation password: %s", err)
	}

	if passphrase != passphraseConfirm {
		fmt.Println()
		return "", errors.New("passphrase mismatch!")
	}
	return passphrase, nil

}
