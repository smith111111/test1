// This code was autogenerated from push/push.proto, do not edit.
package proto

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-rpc/nrpc"
)

// PushP2PServer is the interface that providers of the service
// PushP2P should implement.
type PushP2PServer interface {
	Logout(ctx context.Context, req LogoutReq) (resp LogoutResp, err error)
	SendMsg(ctx context.Context, req SendMsgReq) (resp SendMsgResp, err error)
}

// PushP2PHandler provides a NATS subscription handler that can serve a
// subscription using a given PushP2PServer implementation.
type PushP2PHandler struct {
	ctx     context.Context
	workers *nrpc.WorkerPool
	nc      nrpc.NatsConn
	server  PushP2PServer
}

func NewPushP2PHandler(ctx context.Context, nc nrpc.NatsConn, s PushP2PServer) *PushP2PHandler {
	return &PushP2PHandler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func NewPushP2PConcurrentHandler(workers *nrpc.WorkerPool, nc nrpc.NatsConn, s PushP2PServer) *PushP2PHandler {
	return &PushP2PHandler{
		workers: workers,
		nc:      nc,
		server:  s,
	}
}

func (h *PushP2PHandler) Subject() string {
	return "proto.PushP2P.>"
}

func (h *PushP2PHandler) Handler(msg *nats.Msg) {
	var ctx context.Context
	if h.workers != nil {
		ctx = h.workers.Context
	} else {
		ctx = h.ctx
	}
	request := nrpc.NewRequest(ctx, h.nc, msg.Subject, msg.Reply)
	// extract method name & encoding from subject
	_, _, name, tail, err := nrpc.ParseSubject(
		"proto", 0, "PushP2P", 0, msg.Subject)
	if err != nil {
		log.Printf("PushP2PHanlder: PushP2P subject parsing failed: %v", err)
		return
	}

	request.MethodName = name
	request.SubjectTail = tail

	// call handler and form response
	var immediateError *nrpc.Error
	switch name {
	case "Logout":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("LogoutHanlder: Logout subject parsing failed: %v", err)
			break
		}
		var req LogoutReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("LogoutHandler: Logout request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.Logout(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	case "SendMsg":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("SendMsgHanlder: SendMsg subject parsing failed: %v", err)
			break
		}
		var req SendMsgReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("SendMsgHandler: SendMsg request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.SendMsg(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	default:
		log.Printf("PushP2PHandler: unknown name %q", name)
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
			log.Printf("PushP2PHandler: PushP2P handler failed to publish the response: %s", err)
		}
	} else {
	}
}

type PushP2PClient struct {
	nc      nrpc.NatsConn
	PkgSubject string
	Subject string
	Encoding string
	Timeout time.Duration
}

func NewPushP2PClient(nc nrpc.NatsConn) *PushP2PClient {
	return &PushP2PClient{
		nc:      nc,
		PkgSubject: "proto",
		Subject: "PushP2P",
		Encoding: "protobuf",
		Timeout: 5 * time.Second,
	}
}

func (c *PushP2PClient) Logout(req LogoutReq) (resp LogoutResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "Logout"

	// call
	err = nrpc.Call(&req, &resp, c.nc, subject, c.Encoding, c.Timeout)
	if err != nil {
		return // already logged
	}

	return
}

func (c *PushP2PClient) SendMsg(req SendMsgReq) (resp SendMsgResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "SendMsg"

	// call
	err = nrpc.Call(&req, &resp, c.nc, subject, c.Encoding, c.Timeout)
	if err != nil {
		return // already logged
	}

	return
}

// PushP2MServer is the interface that providers of the service
// PushP2M should implement.
type PushP2MServer interface {
	SyncDeviceInfo(ctx context.Context, req DeviceInfoReq) (resp DeviceInfoResp, err error)
}

// PushP2MHandler provides a NATS subscription handler that can serve a
// subscription using a given PushP2MServer implementation.
type PushP2MHandler struct {
	ctx     context.Context
	workers *nrpc.WorkerPool
	nc      nrpc.NatsConn
	server  PushP2MServer
}

func NewPushP2MHandler(ctx context.Context, nc nrpc.NatsConn, s PushP2MServer) *PushP2MHandler {
	return &PushP2MHandler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func NewPushP2MConcurrentHandler(workers *nrpc.WorkerPool, nc nrpc.NatsConn, s PushP2MServer) *PushP2MHandler {
	return &PushP2MHandler{
		workers: workers,
		nc:      nc,
		server:  s,
	}
}

func (h *PushP2MHandler) Subject() string {
	return "proto.PushP2M.>"
}

func (h *PushP2MHandler) Handler(msg *nats.Msg) {
	var ctx context.Context
	if h.workers != nil {
		ctx = h.workers.Context
	} else {
		ctx = h.ctx
	}
	request := nrpc.NewRequest(ctx, h.nc, msg.Subject, msg.Reply)
	// extract method name & encoding from subject
	_, _, name, tail, err := nrpc.ParseSubject(
		"proto", 0, "PushP2M", 0, msg.Subject)
	if err != nil {
		log.Printf("PushP2MHanlder: PushP2M subject parsing failed: %v", err)
		return
	}

	request.MethodName = name
	request.SubjectTail = tail

	// call handler and form response
	var immediateError *nrpc.Error
	switch name {
	case "SyncDeviceInfo":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("SyncDeviceInfoHanlder: SyncDeviceInfo subject parsing failed: %v", err)
			break
		}
		var req DeviceInfoReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("SyncDeviceInfoHandler: SyncDeviceInfo request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.SyncDeviceInfo(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	default:
		log.Printf("PushP2MHandler: unknown name %q", name)
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
			log.Printf("PushP2MHandler: PushP2M handler failed to publish the response: %s", err)
		}
	} else {
	}
}

type PushP2MClient struct {
	nc      nrpc.NatsConn
	PkgSubject string
	Subject string
	Encoding string
	Timeout time.Duration
}

func NewPushP2MClient(nc nrpc.NatsConn) *PushP2MClient {
	return &PushP2MClient{
		nc:      nc,
		PkgSubject: "proto",
		Subject: "PushP2M",
		Encoding: "protobuf",
		Timeout: 5 * time.Second,
	}
}

func (c *PushP2MClient) SyncDeviceInfo(req DeviceInfoReq) (resp DeviceInfoResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "SyncDeviceInfo"

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
	PushP2P *PushP2PClient
	PushP2M *PushP2MClient
}

func NewClient(nc nrpc.NatsConn) *Client {
	c := Client{
		nc: nc,
		defaultEncoding: "protobuf",
		defaultTimeout: 5*time.Second,
		pkgSubject: "proto",
	}
	c.PushP2P = NewPushP2PClient(nc)
	c.PushP2M = NewPushP2MClient(nc)
	return &c
}

func (c *Client) SetEncoding(encoding string) {
	c.defaultEncoding = encoding
	if c.PushP2P != nil {
		c.PushP2P.Encoding = encoding
	}
	if c.PushP2M != nil {
		c.PushP2M.Encoding = encoding
	}
}

func (c *Client) SetTimeout(t time.Duration) {
	c.defaultTimeout = t
	if c.PushP2P != nil {
		c.PushP2P.Timeout = t
	}
	if c.PushP2M != nil {
		c.PushP2M.Timeout = t
	}
}