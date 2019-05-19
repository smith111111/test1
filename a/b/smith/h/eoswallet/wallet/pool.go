package wallet

import (
	"errors"
	"github.com/eoscanada/eos-go"
	"sync"
)

// ApiPool is an implementation of the eos.API interface which will handle server failure, rotate servers, and retry API requests.
type ApiPool struct {
	apiEndpoints  []string
	apiCache      []*eos.API
	activeServer     int
	rotationMutex    sync.RWMutex
}

func (p *ApiPool) newMaximumTryEnumerator() *maxTryEnum {
	return &maxTryEnum{max: len(p.apiEndpoints), attempts: 0}
}

type maxTryEnum struct{ max, attempts int }

func (m *maxTryEnum) next() bool {
	var now = m.attempts
	m.attempts++
	return now <= m.max
}

func (p *ApiPool) currentApi() *eos.API {
	p.rotationMutex.RLock()
	defer p.rotationMutex.RUnlock()
	return p.apiCache[p.activeServer]
}

// NewApiPool instantiates a new ApiPool object with the given server APIs
func NewApiPool(endpoints []string) (*ApiPool, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("no client endpoints provided")
	}
	var (
		apiCache = make([]*eos.API, len(endpoints))
		pool        = &ApiPool{
			apiCache:     apiCache,
			apiEndpoints: endpoints,
		}
	)
	for i, apiUrl := range endpoints {
		api := eos.New(apiUrl)
		pool.apiCache[i] = api
	}
	return pool, nil
}

// Start will attempt to connect to the first eos.API server. If it fails to
// connect it will rotate through the servers to try to find one that works.
func (p *ApiPool) Start() error {
	for e := p.newMaximumTryEnumerator(); e.next(); {
		if err := p.rotateAndStartNextApi(); err != nil {
			// Log.Errorf("failed start: %s", err)
			continue
		}
		return nil
	}
	// Log.Errorf("all servers failed to start")
	return errors.New("all eos.API servers failed to start")
}

// rotateAndStartNextApi attempts to start the
// next api's connection. If an error is returned, it can be assumed that new
// api could not start and rotateAndStartNextApi needs to be retried. The caller of this
// method should track the retry attempts so as to not repeat indefinitely.
func (p *ApiPool) rotateAndStartNextApi() error {
	// Signal rotation and wait for connections to drain
	p.rotationMutex.Lock()
	defer p.rotationMutex.Unlock()

	p.activeServer = (p.activeServer + 1) % len(p.apiCache)
	nextApi := p.apiCache[p.activeServer]

	//Log.Infof("trying server %s...", p.apiEndpoints[p.activeServer])
	// Should be first connection signal, ensure rotation isn't triggered elsewhere
	if _, err := nextApi.GetInfo(); err != nil {
		return err
	}
	return nil
}


