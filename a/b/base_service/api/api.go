package api

import (
	exchangeRateApi "galaxyotc/common/service/exchangerate_service"
	"github.com/nats-io/go-nats"
)

var (
	ExchangerateApi *exchangeRateApi.Client
)

func Init(nc *nats.Conn) {
	ExchangerateApi = exchangeRateApi.NewClient(nc)
}