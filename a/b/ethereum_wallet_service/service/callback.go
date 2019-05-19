package service

import (
	"galaxyotc/common/log"

	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/ethereum_wallet"
	"galaxyotc/gc_services/ethereum_wallet_service/api"
	"github.com/golang/protobuf/proto"
)

// 序列化
func Marshal(transaction wi.EthereumTransactionCallback) ([]byte, error) {
	// 交易信息
	transactionCallback := &pb.TransactionCallback{
		IsDeposit: *proto.Bool(transaction.IsDeposit),
		Txid: *proto.String(transaction.Txid),
		From: *proto.String(transaction.From),
		To: *proto.String(transaction.To),
		Contract: *proto.String(transaction.Contract),
		Gas: *proto.Uint64(transaction.Gas),
		GasPrice: *proto.String(transaction.GasPrice),
		Status: *proto.Uint64(transaction.Status),
		Height: *proto.Uint64(transaction.Height),
		Value: *proto.String(transaction.Value),
		BlockTime: *proto.Int64(transaction.BlockTime.Unix()),
	}

	return proto.Marshal(transactionCallback)
}

// 回调处理，判断是充值操作回调还是提币操作回调
func EthereumCallback(transaction wi.EthereumTransactionCallback) {
	transactionByte, err := Marshal(transaction)
	if err != nil {
		log.Errorf("EthereumCallback Error: %s", err.Error())
		return
	}

	if _, err := api.AccountApi.EthereumCallback(transactionByte); err != nil {
		log.Errorf("EthereumCallback Error: %s", err.Error())
		return
	}
}