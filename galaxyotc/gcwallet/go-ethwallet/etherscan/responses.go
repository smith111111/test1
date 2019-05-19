package etherscan

import (
)

/* the ethersan api normal transactions response struct*/
type GetNormalTransactionsResp struct {
	BlockNumber 			string				`json:"blockNumber"`
	TimeStamp 				JSONTime			`json:"timeStamp"`
	Hash 					Hash				`json:"hash"`
	Nonce 					string				`json:"nonce"`
	BlockHash 				Hash				`json:"blockHash"`
	From 					Address				`json:"from"`
	To 						Address				`json:"to"`
	Value 					string				`json:"value"`
	Gas 					string				`json:"gas"`
	GasPrice				string				`json:"gasPrice"`
	GasUsed					string				`json:"gasUsed"`
	IsError					string				`json:"isError"`
	TxreceiptStatus			TransactionStatus	`json:"txreceipt_status"`
	TransactionIndex		string				`json:"transactionIndex"`
	Input 					string				`json:"input"`
	ContractAddress 		Address				`json:"contractAddress"`
	CumulativeGasUsed		string				`json:"cumulativeGasUsed"`
	Confirmations 			string				`json:"confirmations"`
}

/* the etherscan api erc20 token transactions response struct*/
type GetERC20TokenTransactionsResp struct {
	BlockNumber 			string				`json:"blockNumber"`
	TimeStamp 				JSONTime			`json:"timeStamp"`
	Hash 					Hash				`json:"hash"`
	Nonce 					string				`json:"nonce"`
	BlockHash 				Hash				`json:"blockHash"`
	From 					Address				`json:"from"`
	To 						Address				`json:"to"`
	Value 					string				`json:"value"`
	Gas 					string				`json:"gas"`
	GasPrice				string				`json:"gasPrice"`
	GasUsed					string				`json:"gasUsed"`
	TransactionIndex		string				`json:"transactionIndex"`
	Input 					string				`json:"input"`
	ContractAddress 		Address				`json:"contractAddress"`
	CumulativeGasUsed		string				`json:"cumulativeGasUsed"`
	Confirmations 			string				`json:"confirmations"`
	TokenName 				string				`json:"tokenName"`
	TokenSymbol				string				`json:"tokenSymbol"`
	TokenDecimal 			string				`json:"tokenDecimal"`
}

/* the etherscan api get transaction receipt status response struct*/
type CheckTransactionReceiptStatusResp struct {
	Status 	string	`json:"status"`
}