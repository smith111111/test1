package wallet

import "fmt"

// Version represents the eosc command version
var Version string
/*
 in main.go

var version = "dev"

func init() {
	cmd.Version = version
}
*/

// Show the program version
func WalletVersion() {
	fmt.Println("https://github.com/eoscanada/eosc - eosc", Version)
}