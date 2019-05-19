package boltdb

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func createTxnStore(t errorHandler, opts ...func(*storm.Options) error) (*TxnStore, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	initDatabaseTables(db)
	txdb := TxnStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}

	return &txdb, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestTxnsPut(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	tx, err := chainhash.NewHashFromStr("9d9442a075096cacfcaeb7ffce323e6f77d4f837925773eeaeb61f35292e5693")
	if err != nil {
		t.Error(err)
	}
	now := time.Now()

	if err := txdb.Put(tx.String(), 36695405, now, "1000", "EOS", "executed"); err != nil {
		t.Error(err)
	}

	var obj Txn
	if err := txdb.DB.Select(q.Eq("Txid", tx.String())).First(&obj); err != nil  {
		t.Error(err)
	}

	if obj.Txid != tx.String() {
		t.Error("Txns DB put failed")
	}
	if obj.Height != 36695405 {
		t.Error("Txns DB failed to put height")
	}
	if obj.Value != "1000" {
		t.Error("Txns DB failed to put amount")
	}
	if obj.Symbol != "EOS" {
		t.Error("Txns DB failed to put symbol")
	}
	if obj.Status != "executed" {
		t.Error("Txns DB failed to put status")
	}
}

func TestTxnsGet(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	tx, err := chainhash.NewHashFromStr("9d9442a075096cacfcaeb7ffce323e6f77d4f837925773eeaeb61f35292e5693")
	if err != nil {
		t.Error(err)
	}
	now := time.Now()

	if err := txdb.Put(tx.String(), 36695405, now, "1000", "EOS", "executed"); err != nil {
		t.Error(err)
	}

	txn, err := txdb.Get(*tx)
	if err != nil {
		t.Error(err)
	}
	if tx.String() != txn.Txid {
		t.Error("Txn DB get failed")
	}
	if txn.Height != 36695405 {
		t.Error("Txns DB failed to put height")
	}
	if txn.Value != "1000" {
		t.Error("Txns DB failed to put amount")
	}
	if txn.Symbol != "EOS" {
		t.Error("Txns DB failed to put symbol")
	}
	if now.Equal(txn.Timestamp) {
		t.Error("Txn DB failed to return correct time")
	}
	if txn.Status != "executed" {
		t.Error("Txns DB failed to put status")
	}
}

func TestTxnsGetAll(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	tx, err := chainhash.NewHashFromStr("9d9442a075096cacfcaeb7ffce323e6f77d4f837925773eeaeb61f35292e5693")
	if err != nil {
		t.Error(err)
	}
	now := time.Now()

	if err := txdb.Put(tx.String(), 36695405, now, "1000", "EOS", "executed"); err != nil {
		t.Error(err)
	}

	txns, err := txdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	if len(txns) < 1 {
		t.Error("Txns DB get all failed")
	}
}

func TestDeleteTxns(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	tx, err := chainhash.NewHashFromStr("9d9442a075096cacfcaeb7ffce323e6f77d4f837925773eeaeb61f35292e5693")
	if err != nil {
		t.Error(err)
	}
	now := time.Now()

	if err := txdb.Put(tx.String(), 36695405, now, "1000", "EOS", "executed"); err != nil {
		t.Error(err)
	}

	err = txdb.Delete(tx)
	if err != nil {
		t.Error(err)
	}
	txns, err := txdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	for _, txn := range txns {
		if txn.Txid == tx.String() {
			t.Error("Txns DB delete failed")
		}
	}
}

func TestTxnStore_UpdateHeight(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	tx, err := chainhash.NewHashFromStr("9d9442a075096cacfcaeb7ffce323e6f77d4f837925773eeaeb61f35292e5693")
	if err != nil {
		t.Error(err)
	}
	now := time.Now()

	if err := txdb.Put(tx.String(), 36695405, now, "1000", "EOS", "executed"); err != nil {
		t.Error(err)
	}

	err = txdb.UpdateHeight(*tx, -1, time.Now())
	if err != nil {
		t.Error(err)
	}
	txn, err := txdb.Get(*tx)
	if txn.Height != -1 {
		t.Error("Txn DB failed to update height")
	}
}
