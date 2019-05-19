package service

import (
	"galaxyotc/common/log"
	wi "galaxyotc/wallet-interface"
	pb "galaxyotc/common/proto/wallet/eos_wallet"
	"galaxyotc/gc_services/eos_wallet_service/api"
	"github.com/golang/protobuf/proto"
	"fmt"
)

// 序列化
func Marshal(transaction *wi.EosTransactionCallback) ([]byte, error) {
	// 交易信息
	transactionCallback := &pb.TransactionCallback{
		IsDeposit: *proto.Bool(transaction.IsDeposit),
		Txid:      *proto.String(transaction.Txid),
		From:      *proto.String(transaction.From),
		To:        *proto.String(transaction.To),
		Contract:  *proto.String(transaction.Contract),
		Status:    *proto.Uint64(transaction.Status),
		Quantity:  *proto.String(transaction.Quantity),
		Memo:      *proto.String(transaction.Memo),
		BlockTime: *proto.Int64(transaction.BlockTime.Unix()),
	}

	return proto.Marshal(transactionCallback)
}

func EosCallback(p *wi.EosTransactionCallback) {
	fmt.Println("EosCallback: ", p)

	buf, err := Marshal(p)
	if err != nil {
		log.Errorf("EthereumCallback Error: %s", err.Error())
		return
	}

	if _, err := api.AccountApi.EosCallback(buf); err != nil {
		log.Errorf("EosCallback Error: %s", err)
		return
	}
}
