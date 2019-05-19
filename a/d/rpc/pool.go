package rpc

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/ratelimit"
	"galaxyotc/common/errors"
	"galaxyotc/common/log"
	"github.com/nats-io/go-nats"
)

type natsClient struct {
	tg string
	rv reflect.Value
	cc *nats.Conn
	tb *ratelimit.Bucket

	mu        *sync.RWMutex
	connected bool

	int64
}

func (nc *natsClient) Close() {
	if nc.cc == nil {
		return
	}
	nc.cc.Close()
	nc.cc = nil
}

func (nc *natsClient) FirstTake() {
	nc.connected = true
	nc.mu = new(sync.RWMutex)
	nc.tb = ratelimit.NewBucket(time.Second, 100000)
	nc.Take()
}

func (nc *natsClient) Take() time.Duration {
	atomic.AddInt64(&nc.int64, 1)
	return nc.tb.Take(1)
}

func (nc *natsClient) Finish() {
	atomic.AddInt64(&nc.int64, -1)
}

type ClientPool struct {
	lb      *lb
	gen     reflect.Value
	conns   chan *natsClient
	abandon chan *natsClient
}

func NewClientPool(target []string, factory reflect.Value, capacity, maxCap int) *ClientPool {
	if capacity <= 0 || maxCap <= 0 || capacity > maxCap {
		panic(errors.New("invalid/out of range capacity"))
	}
	cp := &ClientPool{
		lb:      NewLB(target),
		gen:     factory,
		conns:   make(chan *natsClient, maxCap*len(target)),
		abandon: make(chan *natsClient, maxCap*len(target)),
	}
	for i := 0; i < capacity*len(target); i++ {
		cp.conns <- &natsClient{}
	}
	go cp.gc()
	return cp
}

func (cp *ClientPool) gc() {
	for nc := range cp.abandon {
		go func(nc *natsClient) {
			t := time.Tick(time.Second)
			for _ = range t {
				if using := atomic.LoadInt64(&nc.int64); using <= 0 || nc.cc == nil {
					log.Debug("closing... ", nc.tg)
					nc.Close()
					break
				}
			}
		}(nc)
	}
}

func (cp *ClientPool) wait(conn *natsClient, d time.Duration, abandon bool) {
	cp.Reconnect(conn)
	if d > time.Millisecond {
		cp.abandon <- conn
		cp.conns <- &natsClient{} // 替换conn
		if len(cp.conns) < cap(cp.conns) {
			cp.conns <- &natsClient{} // 增大连接数
		} else {
			log.Warn(errors.New("reach max capacity").WithData(d))
		}
	} else if abandon {
		cp.abandon <- conn
		cp.conns <- &natsClient{} // 替换conn
	} else {
		cp.conns <- conn
	}
	if d <= 0 {
		return
	}
	time.Sleep(d)
}

// 返回一个Nats链接，如果达到并发限制则等待
func (cp *ClientPool) Get() (*natsClient, error) {
	tg := cp.lb.Get()
	conn := <-cp.conns
	if conn.cc != nil {
		cp.wait(conn, conn.Take(), tg != conn.tg)
		return conn, nil
	}

	var err error
	for i := 0; i < 3; i++ {
		conn.cc, err = nats.Connect(tg, nats.Timeout(5 * time.Second))
		if err == nil {
			conn.tg = tg
			conn.FirstTake()
			conn.rv = reflect.ValueOf(
				cp.gen.Call([]reflect.Value{reflect.ValueOf(conn.cc)})[0].Interface(),
			)
			cp.conns <- conn
			break
		}
		tg = cp.lb.Get()
	}
	if err != nil {
		cp.conns <- &natsClient{}
	}
	return conn, err
}

func (cp *ClientPool) Reconnect(nc *natsClient) {
	nc.mu.RLock()
	if nc.connected {
		nc.mu.RUnlock()
		return
	}
	nc.mu.RUnlock()

	nc.mu.Lock()
	defer nc.mu.Unlock()
	if nc.connected {
		return
	}

	tg := nc.tg
	for {
		cc, err := nats.Connect(tg, nats.Timeout(5 * time.Second))
		if err != nil {
			log.Warn(errors.New("reconnect error").WithCause(err))
			tg = cp.lb.Get()
			continue
		}
		nc.tg = tg
		nc.rv = reflect.ValueOf(
			cp.gen.Call([]reflect.Value{reflect.ValueOf(cc)})[0].Interface(),
		)
		nc.cc.Close()
		nc.cc = cc
		nc.connected = true
		return
	}
}

func (cp *ClientPool) Close(nc *natsClient) {
	nc.Finish()
	nc.connected = false
}