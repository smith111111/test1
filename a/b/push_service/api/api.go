package api

import (
	frontApi "galaxyotc/common/service/front_service"
	pushApi "galaxyotc/common/service/push_service"
	"github.com/nats-io/go-nats"
)

var (
	FrontApi 	*frontApi.Client
	PushApi 	*pushApi.Client
)

func Init(nc *nats.Conn) {
	FrontApi = frontApi.NewClient(nc)
	PushApi = pushApi.NewClient(nc)
}