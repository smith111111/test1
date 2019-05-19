package boltdb

import (
	"encoding/hex"
	"gcwallet/eos-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"sync"
	"time"
)

type TxnStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (t *TxnStore) Put(txid string, height int32, timestamp time.Time, value string, symbol string, status string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	err := t.DB.One("Txid", txid, &obj)
	if err != nil {
		obj = Txn{Txid: txid}
	}

	obj.Height = height
	obj.Timestamp = timestamp.Unix()
	obj.Value = value
	obj.Symbol = symbol
	obj.Status = status

	return t.DB.Save(&obj)
}

func (t *TxnStore) Get(txid chainhash.Hash) (wallet.Txn, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	var txn wallet.Txn
	var obj Txn
	id := hex.EncodeToString(txid[:])
	if err := t.DB.Select(q.Eq("Txid", id)).First(&obj); err != nil  {
		return txn, err
	}

	txn = wallet.Txn{
		Txid:      	obj.Txid,//msgTx.TxHash().String(),
		Height:    	obj.Height,
		Timestamp: 	time.Unix(obj.Timestamp, 0),
		Value: 		obj.Value,
		Symbol: 	obj.Symbol,
		Status: 	obj.Status,
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
			Txid:      	obj.Txid, //msgTx.TxHash().String(),
			Height:    	obj.Height,
			Timestamp: 	time.Unix(obj.Timestamp, 0),
			Value: 		obj.Value,
			Symbol:     obj.Symbol,
			Status: 	obj.Status,
		}
		ret = append(ret, txn)
	}
	return ret, nil
}

func (t *TxnStore) Delete(txid *chainhash.Hash) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.DB.Select(q.Eq("Txid", txid.String())).Delete(&Txn{})
}

func (t *TxnStore) UpdateHeight(txid chainhash.Hash, height int32, timestamp time.Time) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	if err := t.DB.Select(q.Eq("Txid", txid.String())).First(&obj); err != nil  {
		return err
	}
	obj.Height = height
	obj.Timestamp = timestamp.Unix()
	return t.DB.Update(&obj)
}
