package init

import (
	"github.com/spf13/viper"
	"galaxyotc/common/model"
	"galaxyotc/common/config"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/push_service/api"
	"galaxyotc/gc_services/push_service/service"

	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func init() {
	// 通用配置
	config.DefaultViper()
	// 当前服务配置
	config.SpecifyViper("push_service", "toml", utils.GetConfigPath())

	// 初始化日志
	log.Init()

	// 初始化数据库
	model.NewDB(viper.GetString("db.gc_dsn"))
	model.NewRedis()

	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("push_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	// 初始化API客户端
	api.Init(nc)

	// 启动RPC服务
	service.NewService()
	if err := service.PService.Start(serviceName, nc); err != nil {
		log.Fatalf("Running Service Error: %s", err.Error())
	}
}