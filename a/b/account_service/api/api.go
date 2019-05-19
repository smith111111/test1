package api

import (
	ethereumWalletApi "galaxyotc/common/service/wallet_service/ethereum_wallet_service"
	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
	multiWalletApi "galaxyotc/common/service/wallet_service/multi_wallet_service"
	eosWalletApi "galaxyotc/common/service/wallet_service/eos_wallet_service"
	exchangeRateApi "galaxyotc/common/service/exchangerate_service"
	userApi "galaxyotc/common/service/backend_service/user"
	taskApi "galaxyotc/common/service/task_service"
	"github.com/nats-io/go-nats"
)

var (
	EthereumWalletApi *ethereumWalletApi.Client
	PrivateWalletApi *privateWalletApi.Client
	MultiWalletApi *multiWalletApi.Client
	EosWalletApi *eosWalletApi.Client
	ExchangerateApi *exchangeRateApi.Client
	UserApi *userApi.Client
	TaskApi *taskApi.Client

)

func Init(nc *nats.Conn) {
	EthereumWalletApi = ethereumWalletApi.NewClient(nc)
	PrivateWalletApi = privateWalletApi.NewClient(nc)
	MultiWalletApi = multiWalletApi.NewClient(nc)
	EosWalletApi = eosWalletApi.NewClient(nc)
	ExchangerateApi = exchangeRateApi.NewClient(nc)
	UserApi = userApi.NewClient(nc)
	TaskApi = taskApi.NewClient(nc)
}