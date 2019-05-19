package boltdb

import (
	"fmt"
	"gcwallet/btc-wallet-interface"
	"github.com/asdine/storm"
	"log"
	"sync"
	"time"
	"os"
	"errors"
	"path"
)

// This database is mostly just an example implementation used for testing.
// End users are free to user their own database.
type BoltDatastore struct {
	keys           wallet.Keys
	utxos          wallet.Utxos
	stxos          wallet.Stxos
	txns           wallet.Txns
	watchedScripts wallet.WatchedScripts
	DB             *storm.DB
	lock           *sync.RWMutex
}

func Create(datadir, prefix string) (*BoltDatastore, error) {
	//err := utils.EnsurePath(repoPath)
	//if err != nil {
	//	log.Fatalln("Fail to create repo directory")
	//}
	//db, err := storm.Open(path.Join(repoPath, "wallet.db"))
	db, err := storm.Open(path.Join(datadir,".gcwallet", fmt.Sprintf("%s_wallet.db", prefix)))
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	l := new(sync.RWMutex)
	boltStore := &BoltDatastore{
		keys: &KeyStore{
			DB:   db,
			lock: l,
		},
		utxos: &UtxoStore{
			DB:   db,
			lock: l,
		},
		stxos: &StxoStore{
			DB:   db,
			lock: l,
		},
		txns: &TxnStore{
			DB:   db,
			lock: l,
		},
		watchedScripts: &WatchedScriptStore{
			DB:   db,
			lock: l,
		},
		DB:   db,
		lock: l,
	}
	initDatabaseTables(db)
	return boltStore, nil
}


type BoltMultiwalletDatastore struct {
	db map[wallet.CoinType]wallet.Datastore
	sync.Mutex
}

func (m *BoltMultiwalletDatastore) GetDatastoreForWallet(coinType wallet.CoinType) (wallet.Datastore, error) {
	m.Lock()
	defer m.Unlock()
	db, ok := m.db[coinType]
	if !ok {
		return nil, errors.New("Cointype not supported")
	}
	return db, nil
}

func NewBoltMultiwalletDatastore(datadir string) *BoltMultiwalletDatastore {
	db := make(map[wallet.CoinType]wallet.Datastore)
	//var err error
	db[wallet.Bitcoin], _ = Create(datadir,"bitcoin")
	db[wallet.BitcoinCash], _ = Create(datadir,"bitcoincash")
	db[wallet.Zcash], _ = Create(datadir,"zcash")
	db[wallet.Litecoin], _ = Create(datadir,"litecoin")
	//db[wallet.Ethereum], err = Create("ethereum")

	return &BoltMultiwalletDatastore{db: db}
}

// 测试用
type ErrorHandler interface {
	Error(args ...interface{})
}


func CreateForTest(datadir, prefix string , t ErrorHandler, opts ...func(*storm.Options) error) (*BoltDatastore, func()) {
	ethereumDS, err:=Create(datadir, prefix)
	if err != nil {
		t.Error(err)
	}

	return ethereumDS, func() {
		ethereumDS.DB.Close()
		os.Remove(path.Join(datadir,".gcwallet", fmt.Sprintf("%s_wallet.db", prefix)))
	}
}

func (s *BoltDatastore) Keys() wallet.Keys {
	return s.keys
}

func (s *BoltDatastore) Utxos() wallet.Utxos {
	return s.utxos
}

func (s *BoltDatastore) Stxos() wallet.Stxos {
	return s.stxos
}

func (s *BoltDatastore) Txns() wallet.Txns {
	return s.txns
}

func (s *BoltDatastore) WatchedScripts() wallet.WatchedScripts {
	return s.watchedScripts
}

func initDatabaseTables(db *storm.DB) error {
	err := db.Init(&Key{})
	if err != nil {
		return err
	}
	err = db.Init(&Utxo{})
	if err != nil {
		return err
	}
	err = db.Init(&Stxo{})
	if err != nil {
		return err
	}

	err = db.Init(&Txn{})
	if err != nil {
		return err
	}
	err = db.Init(&WatchedScript{})
	if err != nil {
		return err
	}

	return db.Init(&Config{})
}

func (s *BoltDatastore) GetMnemonic() (string, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var config Config
	err := s.DB.One("Key", "mnemonic", &config)
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

func (s *BoltDatastore) SetMnemonic(mnemonic string) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var config Config
	err := s.DB.One("Key", "mnemonic", &config)
	if err != nil {
		config = Config{Key: "mnemonic"}
	}
	config.Value = mnemonic
	return s.DB.Save(&config)
}

func (s *BoltDatastore) GetCreationDate() (time.Time, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	var t time.Time
	var config Config
	err := s.DB.One("Key", "creationDate", &config)
	if err != nil {
		return t, err
	}
	return time.Parse(time.RFC3339, config.Value)
}

func (s *BoltDatastore) SetCreationDate(creationDate time.Time) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var config Config
	err := s.DB.One("Key", "creationDate", &config)
	if err != nil {
		config = Config{Key: "creationDate"}
	}
	config.Value = creationDate.Format(time.RFC3339)
	return s.DB.Save(&config)
}
