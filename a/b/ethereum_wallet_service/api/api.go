package api

import (
	accountApi "galaxyotc/common/service/backend_service/account"
	"github.com/nats-io/go-nats"
)

var (
	AccountApi *accountApi.Client
)

func Init(nc *nats.Conn) {
	AccountApi = accountApi.NewClient(nc)
}