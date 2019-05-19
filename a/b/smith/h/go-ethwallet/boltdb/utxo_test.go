package boltdb

import (
	"bytes"
	"encoding/hex"
	"gcwallet/eth-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func createUtxoStore(t errorHandler, opts ...func(*storm.Options) error) (*UtxoStore, wallet.Utxo, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	initDatabaseTables(db)
	uxdb := UtxoStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}
	sh1, _ := chainhash.NewHashFromStr("e941e1c32b3dd1a68edc3af9f7fe711f35aaca60f758c2dd49561e45ca2c41c0")
	outpoint := wire.NewOutPoint(sh1, 0)
	utxo := wallet.Utxo{
		Op:           *outpoint,
		AtHeight:     300000,
		Value:        "100000000",
		ScriptPubkey: []byte("scriptpubkey"),
		WatchOnly:    false,
	}

	return &uxdb, utxo, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestUtxoPut(t *testing.T) {
	uxdb, utxo, cleanup := createUtxoStore(t)
	defer  cleanup()

	err := uxdb.Put(utxo)
	if err != nil {
		t.Error(err)
	}
	var obj Utxo
	o := utxo.Op.Hash.String() + ":" + strconv.Itoa(int(utxo.Op.Index))
	if err := uxdb.DB.Select(q.Eq("Outpoint", o)).First(&obj); err != nil  {
		t.Error(err)
	}
	if obj.Outpoint != o {
		t.Error("Utxo DB returned wrong outpoint")
	}
	if obj.Value != utxo.Value {
		t.Error("Utxo DB returned wrong value")
	}
	if obj.Height != int(utxo.AtHeight) {
		t.Error("Utxo DB returned wrong height")
	}
	if obj.ScriptPubkey != hex.EncodeToString(utxo.ScriptPubkey) {
		t.Error("Utxo DB returned wrong scriptPubKey")
	}
}

func TestUtxoGetAll(t *testing.T) {
	uxdb, utxo, cleanup := createUtxoStore(t)
	defer  cleanup()

	err := uxdb.Put(utxo)
	if err != nil {
		t.Error(err)
	}
	utxos, err := uxdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	if utxos[0].Op.Hash.String() != utxo.Op.Hash.String() {
		t.Error("Utxo DB returned wrong outpoint hash")
	}
	if utxos[0].Op.Index != utxo.Op.Index {
		t.Error("Utxo DB returned wrong outpoint index")
	}
	if utxos[0].Value != utxo.Value {
		t.Error("Utxo DB returned wrong value")
	}
	if utxos[0].AtHeight != utxo.AtHeight {
		t.Error("Utxo DB returned wrong height")
	}
	if !bytes.Equal(utxos[0].ScriptPubkey, utxo.ScriptPubkey) {
		t.Error("Utxo DB returned wrong scriptPubKey")
	}
}

func TestSetWatchOnlyUtxo(t *testing.T) {
	uxdb, utxo, cleanup := createUtxoStore(t)
	defer  cleanup()

	err := uxdb.Put(utxo)
	if err != nil {
		t.Error(err)
	}
	err = uxdb.SetWatchOnly(utxo)
	if err != nil {
		t.Error(err)
	}
	var obj Utxo
	o := utxo.Op.Hash.String() + ":" + strconv.Itoa(int(utxo.Op.Index))
	if err := uxdb.DB.Select(q.Eq("Outpoint", o)).First(&obj); err != nil  {
		t.Error(err)
	}

	if !obj.WatchOnly {
		t.Error("Utxo freeze failed")
	}
}

func TestDeleteUtxo(t *testing.T) {
	uxdb, utxo, cleanup := createUtxoStore(t)
	defer  cleanup()

	err := uxdb.Put(utxo)
	if err != nil {
		t.Error(err)
	}
	err = uxdb.Delete(utxo)
	if err != nil {
		t.Error(err)
	}
	utxos, err := uxdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	if len(utxos) != 0 {
		t.Error("Utxo DB delete failed")
	}
}
