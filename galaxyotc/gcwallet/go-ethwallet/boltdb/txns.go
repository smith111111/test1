package boltdb

import (
	"gcwallet/eth-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"sync"
	"time"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

type TxnStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (t *TxnStore) Put(txid, value, from, to, gas, status, contract string, height int32, timestamp int64, input string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	if value == ""{
		value = "0"
	}
	err := t.DB.One("Txid", txid, &obj)
	if err != nil {
		obj = Txn{Txid: strings.ToLower(txid)}
	}

	obj.Value = value
	obj.From = strings.ToLower(from)
	obj.To = strings.ToLower(to)
	obj.Gas = gas
	obj.Status = status
	obj.Height = height
	obj.Timestamp = timestamp
	obj.ContractAddress = strings.ToLower(contract)
	obj.Input = input

	return t.DB.Save(&obj)
}

func (t *TxnStore) Get(txid string) (wallet.Txn, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	var txn wallet.Txn
	var obj Txn

	txid = strings.ToLower(txid)
	if err := t.DB.Select(q.And(q.Eq("Txid", txid))).First(&obj); err != nil  {
		return txn, err
	}

	txn = wallet.Txn{
		Txid:      	common.HexToHash(obj.Txid),
		Value:     	obj.Value,
		Height:    	obj.Height,
		From: 		common.HexToAddress(obj.From),
		To:			common.HexToAddress(obj.To),
		Gas:		obj.Gas,
		Status: 	obj.Status,
		Timestamp: 	time.Unix(obj.Timestamp, 0),
		Input: 		obj.Input,
	}
	return txn, nil
}

func (t *TxnStore) GetAll() ([]wallet.Txn, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	var ret []wallet.Txn
	var txns []Txn
	if err := t.DB.All(&txns); err != nil  {
		return ret, err
	}

	for _, obj:= range txns {
		txn := wallet.Txn{
			Txid:      	common.HexToHash(obj.Txid),
			Value:     	obj.Value,
			Height:    	obj.Height,
			From: 		common.HexToAddress(obj.From),
			To:			common.HexToAddress(obj.To),
			Gas:		obj.Gas,
			Status: 	obj.Status,
			Timestamp: time.Unix(obj.Timestamp, 0),
			Input: 		obj.Input,
		}
		ret = append(ret, txn)
	}
	return ret, nil
}

func (t *TxnStore) GetAllByLimit(addr string, transactionType, sort, offset, limit int) ([]wallet.Txn, int32, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	var (
		ret []wallet.Txn
		txns []Txn
	)
	var query storm.Query
	// 是否筛选交易类型, (1, 转入, 2, 转出, 3, 全部)
	addr = strings.ToLower(addr)
	switch transactionType {
	case 1:
		query = t.DB.Select(q.Eq("To", addr))
	case 2:
		query = t.DB.Select(q.Eq("From", addr))
	case 3:
		query = t.DB.Select(q.Or(q.Eq("To", addr), q.Eq("From", addr)))
	}

	// 是否反转排序, (1, DESC 2,ASC)
	if sort == 1 {
		query = query.Reverse()
	}

	// 获取记录总数
	count, _ := query.Count(&Txn{})

	if err := query.Skip(offset).Limit(limit).OrderBy("Timestamp").Find(&txns); err != nil  {
		return ret, 0, err
	}
	for _, obj := range txns {
		txn := wallet.Txn{
			Txid:      	common.HexToHash(obj.Txid),
			Value:     	obj.Value,
			Height:    	obj.Height,
			From: 		common.HexToAddress(obj.From),
			To:			common.HexToAddress(obj.To),
			Gas:		obj.Gas,
			Status: 	obj.Status,
			Timestamp: time.Unix(obj.Timestamp, 0),
			Input: 		obj.Input,
		}
		ret = append(ret, txn)
	}
	return ret, int32(count), nil
}

func (t *TxnStore) Delete(txid string) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.DB.Select(q.Eq("Txid", txid)).Delete(&Txn{})
}

func (t *TxnStore) UpdateHeight(txid string, height int32, timestamp int64) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	if err := t.DB.Select(q.Eq("Txid", txid)).First(&obj); err != nil  {
		return err
	}
	obj.Height = height
	obj.Timestamp = timestamp
	return t.DB.Update(&obj)
}
