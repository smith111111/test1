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
	"strings"
)

func createAccountStore(t errorHandler, opts ...func(*storm.Options) error) (*AccountStore, wallet.Account, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	initDatabaseTables(db)
	as := AccountStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}
	account := wallet.Account{
		Name:          "test11111115",
		PublicKey:     "EOS6uZTdHynhCyph2TqUEzWsy9r1K2wLxpsKxK4VFUz14GyBqW92H",
		Authority:     []string{"owner, active"},
	}

	return &as, account, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestAccountPut(t *testing.T) {
	db, account, cleanup := createAccountStore(t)
	defer  cleanup()

	if err := db.Put(account); err != nil {
		t.Error(err)
	}
	var obj Account
	if err := db.DB.Select(q.Eq("Name", account.Name)).First(&obj); err != nil  {
		t.Error(err)
	}

	t.Logf("account is: %+v", obj)

	if obj.Name != account.Name {
		t.Error("Account DB returned wrong name")
	}
	if obj.PublicKey != account.PublicKey {
		t.Error("Account DB returned wrong public key")
	}
	if obj.Authority != strings.Join(account.Authority, ",") {
		t.Error("Account DB returned wrong authority")
	}
}

func TestAccountGetAll(t *testing.T) {
	db, account, cleanup := createAccountStore(t)
	defer  cleanup()

	err := db.Put(account)
	if err != nil {
		t.Error(err)
	}

	accounts, err := db.GetAll()
	if err != nil {
		t.Error(err)
	}

	t.Logf("accounts[0] is: %+v", accounts[0])

	if accounts[0].Name != account.Name {
		t.Error("Account DB returned wrong name")
	}
	if accounts[0].PublicKey != account.PublicKey {
		t.Error("Account DB returned wrong public key")
	}
	if strings.Join(accounts[0].Authority, ",") != strings.Join(account.Authority, ",") {
		t.Error("Account DB returned wrong authority")
	}
}

func TestDeleteAccount(t *testing.T) {
	db, account, cleanup := createAccountStore(t)
	defer  cleanup()

	err := db.Put(account)
	if err != nil {
		t.Error(err)
	}
	err = db.Delete(account)
	if err != nil {
		t.Error(err)
	}
	accounts, err := db.GetAll()
	if err != nil {
		t.Error(err)
	}
	if len(accounts) != 0 {
		t.Error("Utxo DB delete failed")
	}
}
