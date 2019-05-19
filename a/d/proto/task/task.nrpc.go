// This code was autogenerated from task/task.proto, do not edit.
package proto

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-rpc/nrpc"
)

// TaskServiceServer is the interface that providers of the service
// TaskService should implement.
type TaskServiceServer interface {
	OrderTimeout(ctx context.Context, req OrderTimeoutReq) (resp BoolResp, err error)
	AccountTransfer(ctx context.Context, req AccountTransferReq) (resp BoolResp, err error)
}

// TaskServiceHandler provides a NATS subscription handler that can serve a
// subscription using a given TaskServiceServer implementation.
type TaskServiceHandler struct {
	ctx     context.Context
	workers *nrpc.WorkerPool
	nc      nrpc.NatsConn
	server  TaskServiceServer
}

func NewTaskServiceHandler(ctx context.Context, nc nrpc.NatsConn, s TaskServiceServer) *TaskServiceHandler {
	return &TaskServiceHandler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func NewTaskServiceConcurrentHandler(workers *nrpc.WorkerPool, nc nrpc.NatsConn, s TaskServiceServer) *TaskServiceHandler {
	return &TaskServiceHandler{
		workers: workers,
		nc:      nc,
		server:  s,
	}
}

func (h *TaskServiceHandler) Subject() string {
	return "proto.TaskService.>"
}

func (h *TaskServiceHandler) Handler(msg *nats.Msg) {
	var ctx context.Context
	if h.workers != nil {
		ctx = h.workers.Context
	} else {
		ctx = h.ctx
	}
	request := nrpc.NewRequest(ctx, h.nc, msg.Subject, msg.Reply)
	// extract method name & encoding from subject
	_, _, name, tail, err := nrpc.ParseSubject(
		"proto", 0, "TaskService", 0, msg.Subject)
	if err != nil {
		log.Printf("TaskServiceHanlder: TaskService subject parsing failed: %v", err)
		return
	}

	request.MethodName = name
	request.SubjectTail = tail

	// call handler and form response
	var immediateError *nrpc.Error
	switch name {
	case "OrderTimeout":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("OrderTimeoutHanlder: OrderTimeout subject parsing failed: %v", err)
			break
		}
		var req OrderTimeoutReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("OrderTimeoutHandler: OrderTimeout request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.OrderTimeout(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	case "AccountTransfer":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("AccountTransferHanlder: AccountTransfer subject parsing failed: %v", err)
			break
		}
		var req AccountTransferReq
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("AccountTransferHandler: AccountTransfer request unmarshal failed: %v", err)
			immediateError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func(ctx context.Context)(proto.Message, error){
				innerResp, err := h.server.AccountTransfer(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
		}
	default:
		log.Printf("TaskServiceHandler: unknown name %q", name)
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
			log.Printf("TaskServiceHandler: TaskService handler failed to publish the response: %s", err)
		}
	} else {
	}
}

type TaskServiceClient struct {
	nc      nrpc.NatsConn
	PkgSubject string
	Subject string
	Encoding string
	Timeout time.Duration
}

func NewTaskServiceClient(nc nrpc.NatsConn) *TaskServiceClient {
	return &TaskServiceClient{
		nc:      nc,
		PkgSubject: "proto",
		Subject: "TaskService",
		Encoding: "protobuf",
		Timeout: 5 * time.Second,
	}
}

func (c *TaskServiceClient) OrderTimeout(req OrderTimeoutReq) (resp BoolResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "OrderTimeout"

	// call
	err = nrpc.Call(&req, &resp, c.nc, subject, c.Encoding, c.Timeout)
	if err != nil {
		return // already logged
	}

	return
}

func (c *TaskServiceClient) AccountTransfer(req AccountTransferReq) (resp BoolResp, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "AccountTransfer"

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
	TaskService *TaskServiceClient
}

func NewClient(nc nrpc.NatsConn) *Client {
	c := Client{
		nc: nc,
		defaultEncoding: "protobuf",
		defaultTimeout: 5*time.Second,
		pkgSubject: "proto",
	}
	c.TaskService = NewTaskServiceClient(nc)
	return &c
}

func (c *Client) SetEncoding(encoding string) {
	c.defaultEncoding = encoding
	if c.TaskService != nil {
		c.TaskService.Encoding = encoding
	}
}

func (c *Client) SetTimeout(t time.Duration) {
	c.defaultTimeout = t
	if c.TaskService != nil {
		c.TaskService.Timeout = t
	}
}