// This code was autogenerated from backend/otc/otc.proto, do not edit.
package proto

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-rpc/nrpc"
)

// OTCServiceServer is the interface that providers of the service
// OTCService should implement.
type OTCServiceServer interface {
	OrderTimeoutCallback(ctx context.Context, req OrderTimeoutCallbackReq) (resp BoolResp, err error)
}

// OTCServiceHandler provides a NATS subscription handler that can serve a
// subscription using a given OTCServiceServer implementation.
type OTCServiceHandler struct {
	ctx     context.Context
	workers *nrpc.WorkerPool
	nc      nrpc.NatsConn
	server  OTCServiceServer
}

func NewOTCServiceHandler(ctx context.Context, nc nrpc.NatsConn, s OTCServiceServer) *OTCServiceHandler {
	return &OTCServiceHandler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func NewOTCServiceConcurrentHandler(workers *nrpc.WorkerPool, nc nrpc.NatsConn, s OTCServiceServer) *OTCServiceHandler {
	return &OTCServiceHandler{
		workers: workers,
		nc:      nc,
		server:  s,
	}
}

func (h *OTCServiceHandler) Subject() string {
	return "proto.OTCService.>"
}

func (h *OTCServiceHandler) Handler(msg *nats.Msg) {
	var ctx context.Context
	if h.workers != nil {
		ctx = h.workers.Context
	} else {
		ctx = h.ctx
	}
	request := nrpc.NewRequest(ctx, h.nc, msg.Subject, msg.Reply)
	// extract method name & encoding from subject
	_, _, name, tail, err := nrpc.ParseSubject(
		"proto", 0, "OTCService", 0, msg.Subject)
	if err != nil {
		log.Printf("OTCServiceHanlder: OTCService subject parsing failed: %v", err)
		return
	}

	request.MethodName = name
	request.SubjectTail = tail

	// call handler and form response
	var immediateError *nrpc.Error
	switch name {
	case "OrderTimeoutCallback":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("OrderTimeoutCallbackHanlder: OrderTimeoutCallback subject parsing failed: %v", err)
			break
		}
		var req OrderTimeoutCallbackReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("OrderTimeoutCallbackHandler: OrderTimeoutCallback request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.OrderTimeoutCallback(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	default:
		log.Printf("OTCServiceHandler: unknown name %q", name)
		immediateError = &nrpc.Error{
			Type: nrpc.Error_CLIENT,
			Message: "unknown name: " + name,
		}
	}
	if immediateError == nil {
		if h.workers != nil {
			// Try queuing the request
			if err := h.workers.QueueRequest(request); err != nil {
				log.Printf("nrpc: Error queuing the request: %s", err)
			}
		} else {
			// Run the handler synchronously
			request.RunAndReply()
		}
	}

	if immediateError != nil {
		if err := request.SendReply(nil, immediateError); err != nil {
			log.Printf("OTCServiceHandler: OTCService handler failed to publish the response: %s", err)
		}
	} else {
	}
}

type OTCServiceClient struct {
	nc      nrpc.NatsConn
	PkgSubject string
	Subject string
	Encoding string
	Timeout time.Duration
}

func NewOTCServiceClient(nc nrpc.NatsConn) *OTCServiceClient {
	return &OTCServiceClient{
		nc:      nc,
		PkgSubject: "proto",
		Subject: "OTCService",
		Encoding: "protobuf",
		Timeout: 5 * time.Second,
	}
}

func (c *OTCServiceClient) OrderTimeoutCallback(req OrderTimeoutCallbackReq) (resp BoolResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "OrderTimeoutCallback"

	// call
	err = nrpc.Call(&req, &resp, c.nc, subject, c.Encoding, c.Timeout)
	if err != nil {
		return // already logged
	}

	return
}

type Client struct {
	nc      nrpc.NatsConn
	defaultEncoding string
	defaultTimeout time.Duration
	pkgSubject string
	OTCService *OTCServiceClient
}

func NewClient(nc nrpc.NatsConn) *Client {
	c := Client{
		nc: nc,
		defaultEncoding: "protobuf",
		defaultTimeout: 5*time.Second,
		pkgSubject: "proto",
	}
	c.OTCService = NewOTCServiceClient(nc)
	return &c
}

func (c *Client) SetEncoding(encoding string) {
	c.defaultEncoding = encoding
	if c.OTCService != nil {
		c.OTCService.Encoding = encoding
	}
}

func (c *Client) SetTimeout(t time.Duration) {
	c.defaultTimeout = t
	if c.OTCService != nil {
		c.OTCService.Timeout = t
	}
}