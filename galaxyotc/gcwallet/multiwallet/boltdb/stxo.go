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

type StxoStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (s *StxoStore) Put(stxo wallet.Stxo) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var obj Stxo
	outpoint := stxo.Utxo.Op.Hash.String() + ":" + strconv.Itoa(int(stxo.Utxo.Op.Index))

	err := s.DB.One("Outpoint", outpoint, &obj)
	if err != nil {
		obj = Stxo{Outpoint: outpoint}
	}
	obj.Value = stxo.Utxo.Value
	obj.Height = int(stxo.Utxo.AtHeight)
	obj.ScriptPubkey = hex.EncodeToString(stxo.Utxo.ScriptPubkey)
	obj.WatchOnly = stxo.Utxo.WatchOnly
	obj.SpendHeight = int(stxo.SpendHeight)
	obj.SpendTxid = stxo.SpendTxid.String()

	return s.DB.Save(&obj)
}

func (s *StxoStore) GetAll() ([]wallet.Stxo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var stxos []Stxo
	var ret []wallet.Stxo
	err := s.DB.All(&stxos)
	if err != nil  {
		return ret, err
	}

	for _, stxo:= range stxos {
		s := strings.Split(stxo.Outpoint, ":")
		shaHash, err := chainhash.NewHashFromStr(s[0])
		if err != nil {
			continue
		}
		index, err := strconv.Atoi(s[1])
		if err != nil {
			continue
		}
		scriptBytes, err := hex.DecodeString(stxo.ScriptPubkey)
		if err != nil {
			continue
		}
		spentHash, err := chainhash.NewHashFromStr(stxo.SpendTxid)
		if err != nil {
			continue
		}
		utxo := wallet.Utxo{
			Op:           *wire.NewOutPoint(shaHash, uint32(index)),
			AtHeight:     int32(stxo.Height),
			Value:        stxo.Value,
			ScriptPubkey: scriptBytes,
			WatchOnly:    stxo.WatchOnly,
		}
		ret = append(ret, wallet.Stxo{
			Utxo:        utxo,
			SpendHeight: int32(stxo.SpendHeight),
			SpendTxid:   *spentHash,
		})
	}
	return ret, nil
}

func (s *StxoStore) Delete(stxo wallet.Stxo) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	outpoint := stxo.Utxo.Op.Hash.String() + ":" + strconv.Itoa(int(stxo.Utxo.Op.Index))
	return s.DB.Select(q.Eq("Outpoint", outpoint)).Delete(&Stxo{})
}
