package api

import (
	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
	exchangeRateApi "galaxyotc/common/service/exchangerate_service"
	taskApi "galaxyotc/common/service/task_service"
	pushApi "galaxyotc/common/service/push_service"
	"github.com/nats-io/go-nats"
)

var (
	PrivateWalletApi *privateWalletApi.Client
	ExchangerateApi *exchangeRateApi.Client
	TaskApi *taskApi.Client
	PushApi *pushApi.Client
)

func Init(nc *nats.Conn) {
	PrivateWalletApi = privateWalletApi.NewClient(nc)
	ExchangerateApi = exchangeRateApi.NewClient(nc)
	TaskApi = taskApi.NewClient(nc)
	PushApi = pushApi.NewClient(nc)
}