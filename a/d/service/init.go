package service

import (
	"github.com/nats-io/go-nats"
	"galaxyotc/common/log"
	"time"
	"sync"
)

var (
	nc *nats.Conn
)

func NewAntsClient(addr, name string) *nats.Conn {
	var once sync.Once
	once.Do(func() {
		var err error
		nc, err = nats.Connect(addr, nats.Timeout(60 * time.Second),nats.Name(name))
		if err != nil {
			log.Error(err)
			panic(err)
		}
	})

	return nc
}