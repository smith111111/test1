package boltdb

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"gcwallet/eth-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/btcec"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
)


type errorHandler interface {
	Error(args ...interface{})
}

func createKeyStore(t errorHandler, opts ...func(*storm.Options) error) (*KeyStore, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	initDatabaseTables(db)
	ks := KeyStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}

	return &ks, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestGetAll(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	for i := 0; i < 100; i++ {
		b := make([]byte, 32)
		rand.Read(b)
		err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, i})
		if err != nil {
			t.Error(err)
		}
	}
	all, err := ks.GetAll()
	if err != nil || len(all) != 100 {
		t.Error("Failed to fetch all keys")
	}
}

func TestPutKey(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	b := make([]byte, 32)
	err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, 0})
	if err != nil {
		t.Error(err)
	}
	var key Key
	err = ks.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(b))).First(&key)
	if err != nil  {
		t.Error(err)
	}

	if key.ScriptAddress != hex.EncodeToString(b) {
		t.Errorf(`Expected %s got %s`, hex.EncodeToString(b), key.ScriptAddress)
	}
	if key.Purpose != 0 {
		t.Errorf(`Expected 0 got %d`, key.Purpose)
	}
	if key.KeyIndex != 0 {
		t.Errorf(`Expected 0 got %d`, key.KeyIndex)
	}
	if key.Used {
		t.Errorf(`Expected false got %v`, key.Used)
	}
}

func TestKeysDB_GetImported(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	prvKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error(err)
	}
	err = ks.ImportKey([]byte("fsdfa"), prvKey)
	if err != nil {
		t.Error(err)
	}

	keys, err := ks.GetImported()
	if err != nil {
		t.Error(err)
	}
	if len(keys) != 1 {
		t.Error("Failed to return imported key")
	}
	if !bytes.Equal(prvKey.Serialize(), keys[0].Serialize()) {
		t.Error("Returned incorrect key")
	}
}

func TestImportKey(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	prvKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error(err)
	}
	var b []byte
	for i := 0; i < 32; i++ {
		b = append(b, 0xff)
	}
	err = ks.ImportKey(b, prvKey)
	if err != nil {
		t.Error(err)
	}

	var key Key
	err = ks.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(b))).First(&key)
	if err != nil  {
		t.Error(err)
	}

	if key.ScriptAddress != hex.EncodeToString(b) {
		t.Errorf(`Expected %s got %s`, hex.EncodeToString(b), key.ScriptAddress)
	}
	if key.Purpose != -1 {
		t.Errorf(`Expected -1 got %d`, key.Purpose)
	}
	if !key.Used {
		t.Errorf(`Expected true got %v`, key.Used)
	}
	keyBytes, err := hex.DecodeString(key.Key)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(prvKey.Serialize(), keyBytes) {
		t.Errorf(`Expected %s got %s`, hex.EncodeToString(b), hex.EncodeToString(keyBytes))
	}
}

func TestPutDuplicateKey(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	b := make([]byte, 32)
	ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, 0})
	err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, 0})
	if err == nil {
		t.Error("Expected duplicate key error")
	}
}

func TestMarkKeyAsUsed(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	b := make([]byte, 33)
	err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, 0})
	if err != nil {
		t.Error(err)
	}
	err = ks.MarkKeyAsUsed(b)
	if err != nil {
		t.Error(err)
	}

	var key Key
	err = ks.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(b))).First(&key)
	if err != nil  {
		t.Error(err)
	}

	if !key.Used {
		t.Errorf(`Expected true got %v`, key.Used)
	}
}

func TestGetLastKeyIndex(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	var last []byte
	for i := 0; i < 100; i++ {
		b := make([]byte, 32)
		rand.Read(b)
		err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, i})
		if err != nil {
			t.Error(err)
		}
		last = b
	}
	idx, used, err := ks.GetLastKeyIndex(wallet.EXTERNAL)
	if err != nil || idx != 99 || used != false {
		t.Error("Failed to fetch correct last index")
	}
	ks.MarkKeyAsUsed(last)
	_, used, err = ks.GetLastKeyIndex(wallet.EXTERNAL)
	if err != nil || used != true {
		t.Error("Failed to fetch correct last index")
	}
}

func TestGetPathForKey(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	b := make([]byte, 32)
	rand.Read(b)
	err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, 15})
	if err != nil {
		t.Error(err)
	}
	path, err := ks.GetPathForKey(b)
	if err != nil {
		t.Error(err)
	}
	if path.Index != 15 || path.Purpose != wallet.EXTERNAL {
		t.Error("Returned incorrect key path")
	}
}

func TestGetKey(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	prvKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		t.Error(err)
	}
	var b []byte
	for i := 0; i < 32; i++ {
		b = append(b, 0xee)
	}
	err = ks.ImportKey(b, prvKey)
	if err != nil {
		t.Error(err)
	}
	k, err := ks.GetKey(b)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(prvKey.Serialize(), k.Serialize()) {
		t.Error("Failed to return imported key")
	}
}

func TestKeyNotFound(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	b := make([]byte, 32)
	rand.Read(b)
	_, err := ks.GetPathForKey(b)
	if err == nil {
		t.Error("Return key when it shouldn't have")
	}
}

func TestGetUnsed(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	for i := 0; i < 100; i++ {
		b := make([]byte, 32)
		rand.Read(b)
		err := ks.Put(b, wallet.KeyPath{wallet.INTERNAL, i})
		if err != nil {
			t.Error(err)
		}
	}
	idx, err := ks.GetUnused(wallet.INTERNAL)
	if err != nil {
		t.Error("Failed to fetch correct unused")
	}
	if len(idx) != 100 {
		t.Error("Failed to fetch correct unused")
	}
}

func TestGetLookaheadWindows(t *testing.T) {
	ks, cleanup := createKeyStore(t)
	defer  cleanup()

	for i := 0; i < 100; i++ {
		b := make([]byte, 32)
		rand.Read(b)
		err := ks.Put(b, wallet.KeyPath{wallet.EXTERNAL, i})
		if err != nil {
			t.Error(err)
		}
		if i < 50 {
			ks.MarkKeyAsUsed(b)
		}
		b = make([]byte, 32)
		rand.Read(b)
		err = ks.Put(b, wallet.KeyPath{wallet.INTERNAL, i})
		if err != nil {
			t.Error(err)
		}
		if i < 50 {
			ks.MarkKeyAsUsed(b)
		}
	}
	windows := ks.GetLookaheadWindows()
	if windows[wallet.EXTERNAL] != 50 || windows[wallet.INTERNAL] != 50 {
		t.Error("Fetched incorrect lookahead windows")
	}

}
