package boltdb

import (
	"bytes"
	"encoding/hex"
	"github.com/asdine/storm"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func createWatchedScriptStore(t errorHandler, opts ...func(*storm.Options) error) (*WatchedScriptStore, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "galaxyotc")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "wallet.db"), opts...)
	if err != nil {
		t.Error(err)
	}

	initDatabaseTables(db)
	wsdb := WatchedScriptStore{
		DB:   db,
		lock: new(sync.RWMutex),
	}

	return &wsdb, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestWatchedScriptsDB_Put(t *testing.T) {
	wsdb, cleanup := createWatchedScriptStore(t)
	defer  cleanup()

	err := wsdb.Put([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	var obj WatchedScript
	if err := wsdb.DB.Select().First(&obj); err != nil  {
		t.Error(err)
	}

	if hex.EncodeToString([]byte("test")) != obj.ScriptPubkey {
		t.Error("Failed to inserted watched script into DB")
	}
}

func TestWatchedScriptsDB_GetAll(t *testing.T) {
	wsdb, cleanup := createWatchedScriptStore(t)
	defer  cleanup()

	err := wsdb.Put([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	err = wsdb.Put([]byte("test2"))
	if err != nil {
		t.Error(err)
	}
	scripts, err := wsdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	if len(scripts) != 2 {
		t.Error("Returned incorrect number of watched scripts")
	}
	if !bytes.Equal(scripts[0], []byte("test")) {
		t.Error("Returned incorrect watched script")
	}
	if !bytes.Equal(scripts[1], []byte("test2")) {
		t.Error("Returned incorrect watched script")
	}
}

func TestWatchedScriptsDB_Delete(t *testing.T) {
	wsdb, cleanup := createWatchedScriptStore(t)
	defer  cleanup()

	err := wsdb.Put([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	err = wsdb.Delete([]byte("test"))
	if err != nil {
		t.Error(err)
	}
	scripts, err := wsdb.GetAll()
	if err != nil {
		t.Error(err)
	}
	for _, script := range scripts {
		if bytes.Equal(script, []byte("test")) {
			t.Error("Failed to delete watched script")
		}
	}
}
