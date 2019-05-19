package init

import (
	"github.com/spf13/viper"
	"galaxyotc/common/config"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/exchangerate_service/service"

	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func init() {
	// 加载服务配置
	config.SpecifyViper("exchangerate_service", "toml", utils.GetConfigPath())

	// 初始化日志
	log.Init()

	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("exchangerate_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	// 启动RPC服务
	s := service.NewService()
	if err := s.Start(serviceName, nc); err != nil {
		log.Fatalf("Running Service Error: %s", err.Error())
	}
}