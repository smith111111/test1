package boltdb

import (
	"encoding/hex"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"sync"
)

type WatchedScriptStore struct {
	DB   *storm.DB
	lock *sync.RWMutex
}

func (w *WatchedScriptStore) Put(scriptPubKey []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	var obj WatchedScript
	err := w.DB.One("ScriptPubkey", hex.EncodeToString(scriptPubKey), &obj)
	if err != nil {
		obj = WatchedScript{ScriptPubkey: hex.EncodeToString(scriptPubKey)}
	}
	obj.ScriptPubkey=hex.EncodeToString(scriptPubKey)
	return w.DB.Save(&obj)
}

func (w *WatchedScriptStore) GetAll() ([][]byte, error) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	var ret [][]byte
	var watchedScripts []WatchedScript
	if err := w.DB.All(&watchedScripts); err != nil  {
		if err==storm.ErrNotFound{
			return ret, nil
		}
		return ret, err
	}

	for _, obj := range watchedScripts {
		scriptPubKey, err := hex.DecodeString(obj.ScriptPubkey)
		if err != nil {
			continue
		}
		ret = append(ret, scriptPubKey)
	}
	return ret, nil
}

func (w *WatchedScriptStore) Delete(scriptPubKey []byte) error {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.DB.Select(q.Eq("ScriptPubkey", hex.EncodeToString(scriptPubKey))).Delete(&WatchedScript{})
}
