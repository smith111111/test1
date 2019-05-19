package service

import (
	"galaxyotc/common/log"

	bwi "galaxyotc/btc-wallet-interface"
	pb "galaxyotc/common/proto/wallet/multi_wallet"
	"galaxyotc/gc_services/multi_wallet_service/api"
	"github.com/golang/protobuf/proto"
)

// 序列化
func Marshal(transaction bwi.TransactionCallback) ([]byte, error) {
	// 交易输出
	outputs := []*pb.TransactionOutput{}
	for _, output := range transaction.Outputs {
		outputs = append(outputs, &pb.TransactionOutput{
			Address: *proto.String(output.Address.String()),
			Value: *proto.Int64(output.Value),
			Index: *proto.Uint32(output.Index),
		})
	}

	// 交易输入
	inputs := []*pb.TransactionInput{}
	for _, input := range transaction.Inputs {
		inputs = append(inputs, &pb.TransactionInput{
			OutpointHash: input.OutpointHash,
			OutpointIndex: *proto.Uint32(input.OutpointIndex),
			LinkedAddress: *proto.String(input.LinkedAddress.String()),
			Value: *proto.Int64(input.Value),
		})
	}

	// 交易信息
	transactionCallback := &pb.TransactionCallback{
		Txid: *proto.String(transaction.Txid),
		Outputs: outputs,
		Inputs: inputs,
		Height: *proto.Int32(transaction.Height),
		Timestamp: *proto.Int64(transaction.Timestamp.Unix()),
		Value: *proto.Int64(transaction.Value),
		WatchOnly: *proto.Bool(transaction.WatchOnly),
		BlockTime: *proto.Int64(transaction.BlockTime.Unix()),
	}

	return proto.Marshal(transactionCallback)
}

// 回调处理
func MultiCallback(transaction bwi.TransactionCallback) {
	transactionByte, err := Marshal(transaction)
	if err != nil {
		log.Errorf("MultiCallback Error: %s", err.Error())
		return
	}

	if _, err := api.AccountApi.MultiCallback(transactionByte); err !=nil {
		log.Errorf("MultiCallback Error: %s", err.Error())
		return
	}
}