package service

import (
	"testing"
	wi "galaxyotc/wallet-interface"
	"time"
	"galaxyotc/gc_services/ethereum_wallet_service/api"
	"galaxyotc/common/config"
	"github.com/spf13/viper"
	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func TInit() {
	config.SpecifyViper("ethereum_wallet_service", "toml", utils.GetConfigPath())
	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("ethereum_wallet_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)
	api.Init(nc)
}

func TestEthereumCallback(t *testing.T) {
	TInit()

	transaction := wi.EthereumTransactionCallback {
		IsDeposit: true,
		Txid: "0xb134264a62548ec52a5508683a0e6ee457fbb1c239c5ac89cd13d47c275ece68",
		From: "0x2b2d78acd231cd9c3de22d2dc8ecb7ff8b3bb0f4",
		To: "0xc29dee180b4536bc68316aca247c29d37b74b7b2",
		Contract: "",
		Gas: 21000,
		GasPrice: "1000000000",
		Status: 1,
		Height: 3567267,
		Value: "100000000000000000",
		BlockTime: time.Unix(1545638400, 0),
	}

	EthereumCallback(transaction)
}