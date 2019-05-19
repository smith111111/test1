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
	"github.com/ethereum/go-ethereum/common"
	"strings"
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

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
		t.Error(err)
	}
	var obj Txn
	if err := txdb.DB.Select(q.Eq("Txid", txHex.String())).First(&obj); err != nil  {
		t.Error(err)
	}

	if obj.Txid != txHex.String() {
		t.Error("Txns DB put failed")
	}
	if obj.Value != "1000000000000000000" {
		t.Error("Txns DB failed to put value")
	}
	if obj.From != "0xd3a399c155243c13242addb0200c7ff37e2cf553" {
		t.Error("Txns DB failed to put from")
	}
	if obj.To != "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef" {
		t.Error("Txns DB failed to put to")
	}
}

func TestTxnsGet(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
		t.Error(err)
	}

	txn, err := txdb.Get(txHex.String())
	if err != nil {
		t.Error(err)
	}
	if txHex.String() != txn.Txid.String() {
		t.Error("Txn DB get failed")
	}
	if txn.Value != "1000000000000000000" {
		t.Error("Txns DB failed to put value")
	}
	if strings.ToLower(txn.From.String()) != "0xd3a399c155243c13242addb0200c7ff37e2cf553" {
		t.Error("Txns DB failed to put from")
	}
	if strings.ToLower(txn.To.String()) != "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef" {
		t.Error("Txns DB failed to put to")
	}
}

func TestTxnsGetAll(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
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

func TestGetAllByLimit(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
		t.Error(err)
	}

	txns, count, err := txdb.GetAllByLimit("0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", 3, 1, 0, 20)
	if err != nil {
		t.Error(err)
	}
	if count == 0 {
		t.Error("Txns DB get all failed")
	}
	if len(txns) < 1 {
		t.Error("Txns DB get all failed")
	}
}

func TestDeleteTxns(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
		t.Error(err)
	}

	if err := txdb.Delete(txHex.String()); err != nil {
		t.Error(err)
	}
	txns, err := txdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	for _, txn := range txns {
		if txn.Txid.String() == txHex.String() {
			t.Error("Txns DB delete failed")
		}
	}
}

func TestTxnStore_UpdateHeight(t *testing.T) {
	txdb, cleanup := createTxnStore(t)
	defer  cleanup()

	txHex := common.HexToHash("0x2993b166f61aaf317cbaeab0029068e8ee2e632b748bb732755938a15cad4090")

	if err := txdb.Put(txHex.String(), "1000000000000000000", "0xd3a399c155243c13242addb0200c7ff37e2cf553", "0x9e802fa8a67896f6c6b3edf69114e5565e3017ef", "21000", "success", "", 0, 1544433727, "0x"); err != nil {
		t.Error(err)
	}

	if err := txdb.UpdateHeight(txHex.String(), -1, time.Now().Unix()); err != nil {
		t.Error(err)
	}
	txn, err := txdb.Get(txHex.String())
	if err != nil {
		t.Error(err)
	}

	if txn.Height != -1 {
		t.Error("Txn DB failed to update height")
	}
}
