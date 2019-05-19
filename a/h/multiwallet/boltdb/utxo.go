package boltdb

import (
	"encoding/hex"
	"gcwallet/btc-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"strconv"
	"strings"
	"sync"
)

type UtxoStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (u *UtxoStore) Put(utxo wallet.Utxo) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	var obj Utxo
	outpoint := utxo.Op.Hash.String() + ":" + strconv.Itoa(int(utxo.Op.Index))

	err := u.DB.One("Outpoint", outpoint, &obj)
	if err != nil {
		obj = Utxo{Outpoint: outpoint}
	}
	obj.Value = utxo.Value
	obj.Height = int(utxo.AtHeight)
	obj.ScriptPubkey = hex.EncodeToString(utxo.ScriptPubkey)
	obj.WatchOnly = utxo.WatchOnly

	return u.DB.Save(&obj)
}

func (u *UtxoStore) GetAll() ([]wallet.Utxo, error) {
	u.lock.RLock()
	defer u.lock.RUnlock()
	var ret []wallet.Utxo
	var utxos []Utxo
	if err := u.DB.All(&utxos); err != nil  {
		return ret, err
	}
	for _, obj := range utxos {
		s := strings.Split(obj.Outpoint, ":")
		shaHash, err := chainhash.NewHashFromStr(s[0])
		if err != nil {
			continue
		}
		index, err := strconv.Atoi(s[1])
		if err != nil {
			continue
		}
		scriptBytes, err := hex.DecodeString(obj.ScriptPubkey)
		if err != nil {
			continue
		}
		ret = append(ret, wallet.Utxo{
			Op:           *wire.NewOutPoint(shaHash, uint32(index)),
			AtHeight:     int32(obj.Height),
			Value:        obj.Value,
			ScriptPubkey: scriptBytes,
			WatchOnly:    obj.WatchOnly,
		})
	}
	return ret, nil
}

func (u *UtxoStore) SetWatchOnly(utxo wallet.Utxo) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	outpoint := utxo.Op.Hash.String() + ":" + strconv.Itoa(int(utxo.Op.Index))

	var obj Utxo
	if err := u.DB.Select(q.Eq("Outpoint", outpoint)).First(&obj); err != nil  {
		return  err
	}
	return u.DB.UpdateField(&obj, "WatchOnly", true)
}

func (u *UtxoStore) Delete(utxo wallet.Utxo) error {
	u.lock.Lock()
	defer u.lock.Unlock()
	outpoint := utxo.Op.Hash.String() + ":" + strconv.Itoa(int(utxo.Op.Index))
	return u.DB.Select(q.Eq("Outpoint", outpoint)).Delete(&Utxo{})
}
