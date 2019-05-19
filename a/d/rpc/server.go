package rpc

/*
import (
	"fmt"
	"io"
	"net"
	"reflect"

	"google.golang.org/grpc"
	"galaxyotc/common/log"
)

type Server struct {
	*grpc.Server

	lis     net.Listener
	port    int
	service interface{}
}

func (s *Server) Close() error {
	s.Server.Stop()
	return s.service.(io.Closer).Close()
}

// NewServer 自动完成RPC服务注册，完成RPC的配置并启动RPC服务。
func NewServer(p int, service interface{}, register interface{}, optList ...func(*Server) error) *Server {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := &Server{grpc.NewServer(), lis, p, service}
	// 动态配置RPC服务
	for _, opt := range optList {
		if err := opt(srv); err != nil {
			log.Fatal(err)
		}
	}
	// 注册RPC服务
	// 由于GRPC每个注册服务的接口都指定了类型，所以这里使用reflect完成该注册的调用
	reflect.ValueOf(register).Call([]reflect.Value{reflect.ValueOf(srv.Server), reflect.ValueOf(service)})

	// 指定启动目录，有时调用命令行使需要启动目录
	//os.Chdir(etc.String("system", "chroot"))
	go srv.Serve(lis)
	return srv
}*/