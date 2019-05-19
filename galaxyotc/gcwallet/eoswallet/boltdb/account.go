package boltdb

import (
	"gcwallet/eos-wallet-interface"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"strings"
	"sync"
)

type AccountStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (a *AccountStore) Put(account wallet.Account) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	authority := strings.Join(account.Authority, ",")
	var obj Account
	err := a.DB.One("Name", account.Name, &obj)
	if err != nil {
		obj = Account{Name: account.Name}
	}
	obj.ActivePublicKey = account.ActivePublicKey
	obj.OwnerPublicKey = account.OwnerPublicKey
	obj.Authority = authority

	return a.DB.Save(&obj)
}

func (a *AccountStore) Get(key string) (wallet.Account, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	var obj Account
	err := a.DB.Select(q.Or(q.Eq("ActivePublicKey", key), q.Eq("OwnerPublicKey", key))).First(&obj)
	if err != nil {
		return wallet.Account{}, nil
	}

	account := wallet.Account{
		Name:     		obj.Name,
		ActivePublicKey:    	obj.ActivePublicKey,
		OwnerPublicKey:    	obj.OwnerPublicKey,
		Authority:    	strings.Split(obj.Authority, ","),
	}

	return account, nil
}

func (a *AccountStore) GetAll() ([]wallet.Account, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	var (
		accounts []wallet.Account
		objs []Account
	)

	if err := a.DB.All(&objs); err != nil  {
		return accounts, err
	}
	for _, obj := range objs {
		authority := strings.Split(obj.Authority, ",")

		accounts = append(accounts, wallet.Account{
			Name:     		obj.Name,
			ActivePublicKey:    	obj.ActivePublicKey,
			OwnerPublicKey:    	obj.OwnerPublicKey,
			Authority:    	authority,
		})
	}
	return accounts, nil
}

func (a *AccountStore) Delete(account wallet.Account) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	return a.DB.Select(q.Eq("Name", account.Name)).Delete(&Account{})
}
