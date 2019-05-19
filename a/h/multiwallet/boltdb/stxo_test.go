package boltdb

import (
	"bytes"
	"encoding/hex"
	"gcwallet/btc-wallet-interface"
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

func createStxoStore(t errorHandler, opts ...func(*storm.Options) error) (*StxoStore, wallet.Stxo, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	sxstore := StxoStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}
	sh1, _ := chainhash.NewHashFromStr("e941e1c32b3dd1a68edc3af9f7fe711f35aaca60f758c2dd49561e45ca2c41c0")
	sh2, _ := chainhash.NewHashFromStr("82998e18760a5f6e5573cd789269e7853e3ebaba07a8df0929badd69dc644c5f")
	outpoint := wire.NewOutPoint(sh1, 0)
	utxo := wallet.Utxo{
		Op:           *outpoint,
		AtHeight:     300000,
		Value:        100000000,
		ScriptPubkey: []byte("scriptpubkey"),
		WatchOnly:    false,
	}
	stxo := wallet.Stxo{
		Utxo:        utxo,
		SpendHeight: 300100,
		SpendTxid:   *sh2,
	}

	return &sxstore, stxo, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestStxoPut(t *testing.T) {
	sxstore, stxo, cleanup := createStxoStore(t)
	defer  cleanup()

	err := sxstore.Put(stxo)
	if err != nil {
		t.Error(err)
	}

	var obj Stxo
	o := stxo.Utxo.Op.Hash.String() + ":" + strconv.Itoa(int(stxo.Utxo.Op.Index))
	if err := sxstore.DB.Select(q.Eq("Outpoint", o)).First(&obj); err != nil  {
		t.Error(err)
	}

	if obj.Outpoint != o {
		t.Error("Stxo DB returned wrong outpoint")
	}
	if obj.Value != stxo.Utxo.Value {
		t.Error("Stxo DB returned wrong value")
	}
	if obj.Height != int(stxo.Utxo.AtHeight) {
		t.Error("Stxo DB returned wrong height")
	}
	if obj.ScriptPubkey != hex.EncodeToString(stxo.Utxo.ScriptPubkey) {
		t.Error("Stxo DB returned wrong scriptPubKey")
	}
	if obj.SpendHeight != int(stxo.SpendHeight) {
		t.Error("Stxo DB returned wrong spend height")
	}
	if obj.SpendTxid != stxo.SpendTxid.String() {
		t.Error("Stxo DB returned wrong spend txid")
	}
	if obj.WatchOnly {
		t.Error("Stxo DB returned wrong watch only bool")
	}
}

func TestStxoGetAll(t *testing.T) {
	sxstore, stxo, cleanup := createStxoStore(t)
	defer  cleanup()

	err := sxstore.Put(stxo)
	if err != nil {
		t.Error(err)
	}
	stxos, err := sxstore.GetAll()
	if err != nil {
		t.Error(err)
	}
	if stxos[0].Utxo.Op.Hash.String() != stxo.Utxo.Op.Hash.String() {
		t.Error("Stxo DB returned wrong outpoint hash")
	}
	if stxos[0].Utxo.Op.Index != stxo.Utxo.Op.Index {
		t.Error("Stxo DB returned wrong outpoint index")
	}
	if stxos[0].Utxo.Value != stxo.Utxo.Value {
		t.Error("Stxo DB returned wrong value")
	}
	if stxos[0].Utxo.AtHeight != stxo.Utxo.AtHeight {
		t.Error("Stxo DB returned wrong height")
	}
	if !bytes.Equal(stxos[0].Utxo.ScriptPubkey, stxo.Utxo.ScriptPubkey) {
		t.Error("Stxo DB returned wrong scriptPubKey")
	}
	if stxos[0].SpendHeight != stxo.SpendHeight {
		t.Error("Stxo DB returned wrong spend height")
	}
	if stxos[0].SpendTxid.String() != stxo.SpendTxid.String() {
		t.Error("Stxo DB returned wrong spend txid")
	}
	if stxos[0].Utxo.WatchOnly != stxo.Utxo.WatchOnly {
		t.Error("Stxo DB returned wrong watch only bool")
	}
}

func TestDeleteStxo(t *testing.T) {
	sxstore, stxo, cleanup := createStxoStore(t)
	defer  cleanup()

	err := sxstore.Put(stxo)
	if err != nil {
		t.Error(err)
	}
	err = sxstore.Delete(stxo)
	if err != nil {
		t.Error(err)
	}
	stxos, err := sxstore.GetAll()
	if err != nil {
		t.Error(err)
	}
	if len(stxos) != 0 {
		t.Error("Stxo DB delete failed")
	}
}
