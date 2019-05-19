package api

import (
	otcApi "galaxyotc/common/service/backend_service/otc"
	accountApi "galaxyotc/common/service/backend_service/account"
	"github.com/nats-io/go-nats"
)

var (
	OTCApi *otcApi.Client
	AccountApi *accountApi.Client
)

func Init(nc *nats.Conn) {
	OTCApi = otcApi.NewClient(nc)
	AccountApi = accountApi.NewClient(nc)
}