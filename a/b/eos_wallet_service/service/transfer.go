package service

import (
	"encoding/json"
	"github.com/eoscanada/eos-go"
	"galaxyotc/common/log"
	"galaxyotc/common/utils"
	"github.com/eoscanada/eos-go/token"
	"time"
	"galaxyotc/common/errors"
)

var (
	EosApi *eos.API
)

type ActionTraces struct {
	Receiver string `json:"receiver"`
	Console  string `json:"console"`
	//DataAccess ? `json:"data_access"`
}

type ProcessedData struct {
	Status       string       `json:"status"`
	Id           string       `json:"id"`
	ActionTraces []*ActionTraces `json:"action_traces"`
	//DeferredTransactions ? `json:"deferred_transactions"`
}

type TransferRsp struct {
	StatusCode    string         `json:"StatusCode"`
	TransactionId string         `json:"transaction_id"`
	//Processed     *ProcessedData `json:"processed"`
	BlockId       string         `json:"block_id"`
	BlockNum      int32          `json:"block_num"`
}

func init() {
	keys := []string{"5JihKdESSk8eKz347QsGLmuJL5DDDJC27oH2Z7oDxiTHpXFWU1m", "5Jqm2fdGuH2RmkdaELT7sDNdFLqb6iJy9NnFq97gew8PtP9V8Uv"}

	EosApi = eos.New("https://mainnet1.eoscochain.io")
	EosApi.HttpClient.Timeout = 15 * time.Second

	signer := eos.NewKeyBag()
	for _, key := range keys {
		signer.ImportPrivateKey(key)
	}

	EosApi.SetSigner(signer)
}

func Transfer(from string, to string, amount int64, memo string) (string, error) {
	quantity := eos.NewEOSAsset(amount)

	log.Debug("Transfer, start, from:", from)

	action := token.NewTransfer(eos.AN(from), eos.AN(to), quantity, memo)
	action.Account = eos.AN("eosio.token")
	actions := []*eos.Action{action}
	resp, err := EosApi.SignPushActions(actions...)
	if err != nil {
		log.Error("Transfer, ", err)
		return "", err
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Error("Transfer, ", err)
		return "", err
	}

	log.Debug("Transfer, end, from: ", from, ", data:", utils.RemoveSymbols(string(data), []string{" ", string(0x0A), string(0x0B)}))

	rsp := &TransferRsp{}
	err = json.Unmarshal(data, rsp)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if rsp.StatusCode != "" {
		err = errors.New("rsp.StatusCode != \"\"")
		log.Error(err)
		return "", err
	}

	return rsp.TransactionId, nil
}
