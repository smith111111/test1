package service

import (
	"github.com/gorilla/websocket"
	wi "galaxyotc/wallet-interface"
	"encoding/json"
	"time"
	"github.com/spf13/viper"
	"fmt"
	"strings"
	"strconv"
	"galaxyotc/common/log"
)

var WsClient *websocket.Conn
var ContractAccount string

type AuthorizationInfo struct {
	Actor      string `json:"actor"`
	Permission string `json:"permission"`
}

type DataInfo struct {
	From     string `json:"from"`
	Memo     string `json:"memo"`
	Quantity string `json:"quantity"`
	To       string `json:"to"`
}

type ActionsInfo struct {
	Account       string               `json:"account"`
	Authorization []*AuthorizationInfo `json:"authorization"`
	Data          *DataInfo            `json:"data"`
	HexData       string               `json:"hex_data"`
	Name          string               `json:"name"`
}

type TransferInfo struct {
	TrxId           string         `json:"trx_id"`
	BlockNum        int64          `json:"block_num"`
	GlobalActionSeq int64          `json:"global_action_seq"`
	TrxTimestamp    string         `json:"trx_timestamp"`
	Actions         []*ActionsInfo `json:"actions"`
}

type RspData struct {
	ErrNo   int32  `json:"errno"`
	MsgType string `json:"msg_type"`
	ErrMsg  string `json:"errmsg"`
	Data    []byte `json:"data"`
}

func (this *RspData) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	this.ErrNo, _ = m["errno"].(int32)
	this.MsgType, _ = m["msg_type"].(string)
	this.ErrMsg, _ = m["errmsg"].(string)

	data, err = json.Marshal(m["data"])
	if err != nil {
		return err
	}

	this.Data = data

	return nil
}

type EosWallet struct {
	cb           func(*wi.EosTransactionCallback)
	ownerAddress string
	url          string
	appKey       string
}

func (this *EosWallet) Start() {
	this.ownerAddress = viper.GetString("eos_wallet_service.contract_account")
	log.Debug("OwnerAddress:", this.ownerAddress)
	this.url = viper.GetString("eos_wallet_service.url")
	this.appKey = viper.GetString("eos_wallet_service.app_key")
	url := fmt.Sprintf("%s?apikey=%s", this.url, this.appKey)
	log.Debug("url:", url)

	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	WsClient = client

	go func() {
		for {
			_, message, err := WsClient.ReadMessage()
			if err != nil {
				log.Error(err)
				return
			}

			rspData := &RspData{}
			err = json.Unmarshal(message, rspData)
			if err != nil {
				log.Error(err)
				continue
			}

			if rspData.MsgType != "data" {
				continue
			}

			transferInfo := &TransferInfo{}
			err = json.Unmarshal(rspData.Data, transferInfo)
			if err != nil {
				log.Error(err)
				continue
			}

			p := &wi.EosTransactionCallback{}
			p.Txid = transferInfo.TrxId
			p.BlockTime, _ = time.Parse("2006-01-02T15:04:05.000", transferInfo.TrxTimestamp)

			for _, v := range transferInfo.Actions {
				if v.Account != "eosio.token" || v.Name != "transfer" {
					continue
				}

				if v.Data == nil {
					continue
				}

				if v.Data.From != this.ownerAddress {
					p.IsDeposit = true
				}

				p.From = v.Data.From
				p.To = v.Data.To
				p.Contract = this.ownerAddress

				arr1 := strings.Split(v.Data.Quantity, " ")
				amount, err := strconv.ParseFloat(arr1[0], 64)
				if err != nil {
					continue
				}

				p.Quantity = fmt.Sprintf("%v", int(amount * 10000))

				p.Memo = v.Data.Memo

				this.cb(p)
			}
		}
	}()

	this.sendSubscribeAccount()
}

func (this *EosWallet) EosWithdraw(to string, amount int64, meno string) (string, error) {
	txid, err := Transfer(this.ownerAddress, to, amount, meno)
	return txid, err
}

func (this *EosWallet) AddTransactionListener(cb func(*wi.EosTransactionCallback)) {
	this.cb = cb
}

func (this *EosWallet) sendSubscribeAccount() {
	s := fmt.Sprintf("{\"msg_type\": \"subscribe_account\",\"name\": \"%s\"}", this.ownerAddress)
	err := WsClient.WriteMessage(websocket.TextMessage, []byte(s))
	if err != nil {
		log.Error("write:", err)
		return
	}
}
