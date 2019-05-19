package api

import (
	userApi "galaxyotc/common/service/backend_service/user"
	"github.com/nats-io/go-nats"
)

var (
	UserApi *userApi.Client
)

func Init(nc *nats.Conn) {
	UserApi = userApi.NewClient(nc)
}