package wallet

import (
	"github.com/ethereum/go-ethereum/common"
	"time"
)

type Transaction struct {
	BlockNumber 			int32				`json:"blockNumber"`
	TimeStamp 				time.Time			`json:"timeStamp"`
	Hash 					common.Hash			`json:"hash"`
	Nonce 					string				`json:"nonce"`
	BlockHash 				common.Hash			`json:"blockHash"`
	From 					common.Address		`json:"from"`
	To 						common.Address		`json:"to"`
	Value 					string				`json:"value"`
	Gas 					string				`json:"gas"`
	GasPrice				string				`json:"gasPrice"`
	GasUsed					string				`json:"gasUsed"`
	TxreceiptStatus			string				`json:"txreceipt_status"`
	TransactionIndex		string				`json:"transactionIndex"`
	Input 					string				`json:"input"`
	ContractAddress 		common.Address		`json:"contractAddress"`
	CumulativeGasUsed		string				`json:"cumulativeGasUsed"`
	Confirmations 			string				`json:"confirmations"`
}