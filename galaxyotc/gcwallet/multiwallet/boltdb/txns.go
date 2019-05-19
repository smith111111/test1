package boltdb

import (
	"bytes"
	"gcwallet/btc-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"sync"
	"time"
	"fmt"
)

type TxnStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (t *TxnStore) Put(txn []byte, txid string, value, height int, timestamp time.Time, watchOnly bool) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	err := t.DB.One("Txid", txid, &obj)
	if err != nil {
		fmt.Println("put", txid)
		obj = Txn{Txid: txid}
	}

	obj.Value = int64(value)
	obj.Height = height
	obj.Timestamp = int(timestamp.Unix())
	obj.WatchOnly = watchOnly
	obj.Tx = txn

	return t.DB.Save(&obj)
}

func (t *TxnStore) Get(txid chainhash.Hash) (wallet.Txn, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	var txn wallet.Txn
	var obj Txn

	if err := t.DB.Select(q.Eq("Txid", txid.String())).First(&obj); err != nil  {
		return txn, err
	}
	r := bytes.NewReader(obj.Tx)
	msgTx := wire.NewMsgTx(1)
	msgTx.BtcDecode(r, 1, wire.WitnessEncoding)

	txn = wallet.Txn{
		Txid:      msgTx.TxHash().String(),
		Value:     int64(obj.Value),
		Height:    int32(obj.Height),
		Timestamp: time.Unix(int64(obj.Timestamp), 0),
		WatchOnly: obj.WatchOnly,
		Bytes:     obj.Tx,
	}
	return txn, nil
}

func (t *TxnStore) GetAll(includeWatchOnly bool) ([]wallet.Txn, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	var ret []wallet.Txn
	var txns []Txn
	if err := t.DB.All(&txns); err != nil  {
		return ret, err
	}
	for _, obj := range txns {
		r := bytes.NewReader(obj.Tx)
		msgTx := wire.NewMsgTx(1)
		msgTx.BtcDecode(r, 1, wire.WitnessEncoding)

		if obj.WatchOnly {
			if !includeWatchOnly {
				continue
			}
		}

		txn := wallet.Txn{
			Txid:      msgTx.TxHash().String(),
			Value:     int64(obj.Value),
			Height:    int32(obj.Height),
			Timestamp: time.Unix(int64(obj.Timestamp), 0),
			WatchOnly: obj.WatchOnly,
			Bytes:     obj.Tx,
		}
		ret = append(ret, txn)
	}
	return ret, nil
}

func (t *TxnStore) GetAllByLimit(includeWatchOnly bool, transactionType, sort, offset, limit int) ([]wallet.Txn, int32, error) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	var (
		ret []wallet.Txn
		txns []Txn
	)
	var query storm.Query
	// 是否筛选交易类型, (1, 转入, 2, 转出, 3, 全部)
	switch transactionType {
	case 1:
		query = t.DB.Select(q.Gt("Value", 0))
	case 2:
		query = t.DB.Select(q.Lt("Value", 0))
	case 3:
		query = t.DB.Select(q.Or(q.Gt("Value", 0), q.Lt("Value", 0)))
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
		r := bytes.NewReader(obj.Tx)
		msgTx := wire.NewMsgTx(1)
		msgTx.BtcDecode(r, 1, wire.WitnessEncoding)

		if obj.WatchOnly {
			if !includeWatchOnly {
				continue
			}
		}

		txn := wallet.Txn{
			Txid:      msgTx.TxHash().String(),
			Value:     int64(obj.Value),
			Height:    int32(obj.Height),
			Timestamp: time.Unix(int64(obj.Timestamp), 0),
			WatchOnly: obj.WatchOnly,
			Bytes:     obj.Tx,
		}
		ret = append(ret, txn)
	}
	return ret, int32(count), nil
}

func (t *TxnStore) Delete(txid *chainhash.Hash) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.DB.Select(q.Eq("Txid", txid.String())).Delete(&Txn{})
}

func (t *TxnStore) UpdateHeight(txid chainhash.Hash, height int, timestamp time.Time) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	var obj Txn
	if err := t.DB.Select(q.Eq("Txid", txid.String())).First(&obj); err != nil  {
		return err
	}
	obj.Height = height
	obj.Timestamp = int(timestamp.Unix())
	return t.DB.Update(&obj)
}
