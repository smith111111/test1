package boltdb

import (
	"errors"
	"gcwallet/eos-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"sync"
	"fmt"
)

type KeyStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (k *KeyStore) ImportKey(key wallet.Key) error {
	k.lock.Lock()
	defer k.lock.Unlock()

	if k.hasKey(key.PrivateKey) {
		return fmt.Errorf("the private key already exists")
	}

	var obj = Key{
		Name: key.Name,
		PublicKey: key.PublicKey,
		PrivateKey: key.PrivateKey,
	}

	return k.DB.Save(&obj)
}

func (k *KeyStore) hasKey(privateKey string) bool {
	var obj Key
	err := k.DB.One("PrivateKey", privateKey, &obj)
	if err != nil {
		return false
	}

	return true
}

func (k *KeyStore) HasKey(privateKey string) bool {
	k.lock.Lock()
	defer k.lock.Unlock()

	return k.hasKey(privateKey)
}

func (k *KeyStore) GetPrivateKey(publicKey string) (wallet.Key, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var obj Key
	err := k.DB.Select(q.Eq("PublicKey", publicKey)).First(&obj)
	if err != nil  {
		return wallet.Key{}, errors.New("Key not found")
	}

	key := wallet.Key{
		Name: 		obj.Name,
		PrivateKey: obj.PrivateKey,
		PublicKey: obj.PublicKey,
	}
	return key, nil
}

func (k *KeyStore) GetImported() ([]wallet.Key, error) {
	k.lock.RLock()
	defer k.lock.RUnlock()

	var (
		keys []wallet.Key
		objs []Key
	)

	if err := k.DB.All(&objs); err != nil {
		return keys, err
	}

	for _, obj := range objs {
		keys = append(keys, wallet.Key{
			Name: obj.Name,
			PrivateKey: obj.PrivateKey,
			PublicKey: obj.PublicKey,
		})
	}
	return keys, nil
}

func (a *KeyStore) Delete(key wallet.Key) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.DB.Select(q.Eq("Name", key.Name)).Delete(&Key{})
}
