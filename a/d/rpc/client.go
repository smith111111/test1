package rpc

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"galaxyotc/common/errors"
	"galaxyotc/common/log"
	"galaxyotc/common/utils"

	"github.com/nats-io/go-nats"
)

var (
	eClosing = nats.ErrConnectionClosed.Error()
	eTimeout = nats.ErrTimeout.Error()
)

type methodType struct {
	method    reflect.Method
	argType   reflect.Type
	replyType reflect.Type
}

type service struct {
	mu     sync.Mutex
	name   string
	rcvr   reflect.Value
	method map[string]*methodType
}

func suitableMethods(typ reflect.Type) map[string]*methodType {

	methods := make(map[string]*methodType)

	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		// NOTE：如果存在不合适的方法，调用可能会出现问题
		if method.PkgPath != "" {
			continue
		}
		methods[method.Name] = &methodType{
			method:    method,
			argType:   mtype.In(1),
			replyType: mtype.Out(0),
		}
	}
	return methods
}

var (
	defaultEndpoint   = &endpoint{}
	staticEmptyValue  = reflect.Value{}
	valueOfClientConn = reflect.ValueOf(&nats.Conn{})
)

type endpoint struct {
	mu         sync.RWMutex
	serviceMap map[string]*service
}

func (e *endpoint) register(name string, factory reflect.Value) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, present := e.serviceMap[name]; present {
		log.Fatal("rpc: service already defined: " + name)
	}

	if e.serviceMap == nil {
		e.serviceMap = make(map[string]*service)
	}

	typ := reflect.TypeOf(
		factory.Call([]reflect.Value{valueOfClientConn})[0].Interface(),
	)

	e.serviceMap[name] = &service{
		name:   name,
		method: suitableMethods(typ),
	}
}

func (e *endpoint) getFunc(service, methodName string) (reflect.Value, error) {
	s := e.serviceMap[service]

	if s == nil {
		return staticEmptyValue, errors.New("service not found: " + service)
	}

	mtype := s.method[methodName]

	if mtype == nil {
		return staticEmptyValue, errors.New("method not found: " + methodName)
	}

	return mtype.method.Func, nil
}

type Client struct {
	name       string
	clientPool *ClientPool
}

const rpcClientIdentifier = "rpc_"

func NewClient(name, target string, factory interface{}) *Client {
	name = rpcClientIdentifier + name
	valueOfFactory := reflect.ValueOf(factory)
	defaultEndpoint.register(name, valueOfFactory)
	return &Client{name, NewClientPool(strings.Split(target, ","), valueOfFactory, 10, 50)}
}

func (c *Client) Call(method string, in interface{}) (interface{}, error) {
	//defer metrics.MeasureMethod(c.name + "." + method)()
	f, err := defaultEndpoint.getFunc(c.name, method)
	if err != nil {
		return nil, err
	}

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return nil, errors.NewCode(errors.ErrUnavailable, 3, "当前服务器人数过多,请您稍后再试")
		default:
			rc, e := c.clientPool.Get()
			if e != nil {
				log.Debug(e)
				err = e
				time.Sleep(time.Duration(utils.Random(1000, 0)) * time.Millisecond)
				continue
			}

			r := f.Call([]reflect.Value{rc.rv, reflect.ValueOf(in)})
			if r[1].Interface() == nil {
				rc.Finish()
				return r[0].Interface(), nil
			}

			err = r[1].Interface().(error)
			if strings.Contains(err.Error(), eClosing) || strings.Contains(err.Error(), eTimeout) {
				c.clientPool.Close(rc)
				continue
			}

			rc.Finish()
			return nil, ParseError(err)
		}
	}
}
