package init

import (
	"github.com/spf13/viper"
	"galaxyotc/common/config"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/private_wallet_service/api"
	"galaxyotc/gc_services/private_wallet_service/service"

	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func init() {
	// 加载服务配置
	config.SpecifyViper("private_wallet_service", "toml", utils.GetConfigPath())

	// 初始化日志
	log.Init()

	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("private_wallet_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	// 初始化API客户端
	api.Init(nc)

	// 启动RPC服务
	s := service.NewService()
	if err := s.Start(serviceName, nc); err != nil {
		log.Fatalf("Running Service Error: %s", err.Error())
	}
}