package boltdb

import (
	"gcwallet/eos-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
)


type errorHandler interface {
	Error(args ...interface{})
}

func createKeyStore(t errorHandler, opts ...func(*storm.Options) error) (*KeyStore, wallet.Key, func()) {
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
	key := wallet.Key{
		Name:          "test11111115",
		PrivateKey:    "5KLThSrk2X2DT4mr23zJN3aNNNQhMy5xPiNMbS6sZVtNwQ5Mb4z",
		PublicKey:     "EOS6uZTdHynhCyph2TqUEzWsy9r1K2wLxpsKxK4VFUz14GyBqW92H",
	}

	return &ks, key, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestImportKey(t *testing.T) {
	db, key, cleanup := createKeyStore(t)
	defer  cleanup()

	if err := db.ImportKey(key); err != nil {
		t.Error(err)
	}

	var obj Key
	if err := db.DB.Select(q.Eq("Name", key.Name)).First(&obj); err != nil  {
		t.Error(err)
	}

	t.Logf("key is: %+v", obj)

	if obj.Name != key.Name {
		t.Error("Key DB returned wrong name")
	}
	if obj.PrivateKey != key.PrivateKey {
		t.Error("Key DB returned wrong private key")
	}
	if obj.PublicKey != key.PublicKey {
		t.Error("Key DB returned wrong public key")
	}
}

func TestHasKey(t *testing.T) {
	db, key, cleanup := createKeyStore(t)
	defer  cleanup()

	if err := db.ImportKey(key); err != nil {
		t.Error(err)
	}

	exist := db.HasKey(key.PrivateKey)
	t.Logf("is exist: %t", exist)
	if exist {
		if err := db.DB.Select(q.Eq("Name", key.Name)).Delete(&Key{}); err != nil {
			t.Error(err)
		}

		exist = db.HasKey(key.PrivateKey)
		if exist {
			t.Error("Key DB returned wrong bool")
		}

		t.Logf("is exist: %t", exist)
	}
}

func TestGetPrivateKey(t *testing.T) {
	db, key, cleanup := createKeyStore(t)
	defer  cleanup()

	if err := db.ImportKey(key); err != nil {
		t.Error(err)
	}

	obj, err := db.GetPrivateKey(key.PublicKey)
	if err != nil {
		t.Error(err)
	}

	t.Logf("key is: %+v", obj)

	if obj.Name != key.Name {
		t.Error("Key DB returned wrong name")
	}
	if obj.PrivateKey != key.PrivateKey {
		t.Error("Key DB returned wrong private key")
	}
	if obj.PublicKey != key.PublicKey {
		t.Error("Key DB returned wrong public key")
	}
}

func TestGetImported(t *testing.T) {
	db, key, cleanup := createKeyStore(t)
	defer  cleanup()

	err := db.ImportKey(key)
	if err != nil {
		t.Error(err)
	}

	keys, err := db.GetImported()
	if err != nil {
		t.Error(err)
	}

	t.Logf("keys[0] is: %+v", keys[0])

	if keys[0].Name != key.Name {
		t.Error("Account DB returned wrong name")
	}
	if keys[0].PrivateKey != key.PrivateKey {
		t.Error("Key DB returned wrong private key")
	}
	if keys[0].PublicKey != key.PublicKey {
		t.Error("Key DB returned wrong public key")
	}
}

func TestDeleteKey(t *testing.T) {
	db, key, cleanup := createKeyStore(t)
	defer  cleanup()

	if err := db.ImportKey(key); err != nil {
		t.Error(err)
	}

	if err := db.Delete(key); err != nil {
		t.Error(err)
	}

	keys, err := db.GetImported()
	if err != nil {
		t.Error(err)
	}

	if len(keys) != 0 {
		t.Error("Utxo DB delete failed")
	}
}
