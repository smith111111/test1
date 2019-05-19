package boltdb

import (
	"encoding/hex"
	"errors"
	"gcwallet/eth-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/btcsuite/btcd/btcec"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
)

type KeyStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (k *KeyStore) Put(scriptAddress []byte, keyPath wallet.KeyPath) error {
	k.lock.Lock()
	defer k.lock.Unlock()

	var key = Key{
		ScriptAddress: hex.EncodeToString(scriptAddress),
		Purpose: int(keyPath.Purpose),
		KeyIndex:keyPath.Index,
		Used:false,
	}
	return k.DB.Save(&key)
}

func (k *KeyStore) ImportKey(scriptAddress []byte, privKey *btcec.PrivateKey) error {
	k.lock.Lock()
	defer k.lock.Unlock()

	var key = Key{
		ScriptAddress: hex.EncodeToString(scriptAddress),
		Purpose: -1, // 表示一个导入的私钥
		KeyIndex:int(rand.Uint32()),
		Used:true,
		Key:hex.EncodeToString(privKey.Serialize()),
	}

	return k.DB.Save(&key)
}

func (k *KeyStore) MarkKeyAsUsed(scriptAddress []byte) error {
	k.lock.Lock()
	defer k.lock.Unlock()

	var obj Key
	if err := k.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(scriptAddress))).First(&obj); err != nil  {
		return  err
	}
	return k.DB.UpdateField(&obj, "Used", true)
}

func (k *KeyStore) GetLastKeyIndex(purpose wallet.KeyPurpose) (int, bool, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var key Key
	err := k.DB.Select(q.Eq("Purpose", int(purpose))).OrderBy("ID").Reverse().First(&key)
	if err != nil  {
		return 0, false, err
	}
	return key.KeyIndex, key.Used, nil
}

func (k *KeyStore) GetPathForKey(scriptAddress []byte) (wallet.KeyPath, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var key Key
	err := k.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(scriptAddress)), q.Not(q.Eq("Purpose", -1))).First(&key)
	if err != nil  {
		return wallet.KeyPath{}, errors.New("Key not found")
	}

	p := wallet.KeyPath{
		Purpose: wallet.KeyPurpose(key.Purpose),
		Index:   key.KeyIndex,
	}
	return p, nil
}

func (k *KeyStore) GetKey(scriptAddress []byte) (*btcec.PrivateKey, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var key Key
	err := k.DB.Select(q.Eq("ScriptAddress", hex.EncodeToString(scriptAddress)), q.Eq("Purpose", -1)).First(&key)
	if err != nil  {
		return nil, errors.New("Key not found")
	}

	keyBytes, err := hex.DecodeString(key.Key)
	if err != nil {
		return nil, err
	}
	prvKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
	return prvKey, nil
}

func (k *KeyStore) GetImported() ([]*btcec.PrivateKey, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var keys []Key
	var ret []*btcec.PrivateKey
	err := k.DB.Select(q.Eq("Purpose", -1)).Find(&keys)
	if err != nil  {
		return ret, err
	}

	for _, key:= range keys {
		keyBytes, err := hex.DecodeString(key.Key)
		if err != nil {
			return ret, err
		}
		priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
		ret = append(ret, priv)
	}
	return ret, nil
}

func (k *KeyStore) GetUnused(purpose wallet.KeyPurpose) ([]int, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var keys []Key
	var ret []int
	err := k.DB.Select(q.Eq("Purpose", int(purpose)), q.Eq("Used", false)).OrderBy("ID").Find(&keys)
	if err != nil  {
		return ret, err
	}

	for _, key := range keys {
		ret = append(ret, key.KeyIndex)
	}
	return ret, nil
}

func (k *KeyStore) GetAll() ([]wallet.KeyPath, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var keys []Key
	var ret []wallet.KeyPath
	err := k.DB.All(&keys)
	if err != nil  {
		return ret, err
	}

	for _, key := range keys {
		p := wallet.KeyPath{
			Purpose: wallet.KeyPurpose(key.Purpose),
			Index:   key.KeyIndex,
		}
		ret = append(ret, p)
	}
	return ret, nil
}

func (k *KeyStore) GetLookaheadWindows() map[wallet.KeyPurpose]int {
	k.lock.RLock()
	defer k.lock.RUnlock()

	windows := make(map[wallet.KeyPurpose]int)
	windows[wallet.KeyPurpose(0)] = 0
	windows[wallet.KeyPurpose(1)] = 0
	for i := 0; i < 2; i++ {
		var keys []Key
		err := k.DB.Select(q.Eq("Purpose", i)).OrderBy("ID").Reverse().Find(&keys)
		if err != nil  {
			continue
		}

		var unusedCount int
		for _, key := range keys {
			if !key.Used {
				unusedCount++
			} else {
				break
			}
		}
		purpose := wallet.KeyPurpose(i)
		windows[purpose] = unusedCount
	}
	return windows
}



func CreateKeyStore(t ErrorHandler, opts ...func(*storm.Options) error) (*KeyStore, func()) {
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