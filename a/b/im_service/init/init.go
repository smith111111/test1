package init

import (
	"github.com/spf13/viper"
	"galaxyotc/common/model"
	"galaxyotc/common/config"
	"galaxyotc/common/log"
	"galaxyotc/gc_services/im_service/client"
	"galaxyotc/gc_services/im_service/service"

	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func init() {
	// 通用配置
	config.DefaultViper()
	// 当前服务配置
	config.SpecifyViper("im_service", "toml", utils.GetConfigPath())

	// 初始化日志
	log.Init()

	// 初始化数据库
	model.NewDB(viper.GetString("db.gc_dsn"))
	model.NewRedis()

	// 初始化客户端
	client.Init()

	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("im_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	// 启动RPC服务
	s := service.NewService()
	if err := s.Start(serviceName, nc); err != nil {
		log.Fatalf("Running Service Error: %s", err.Error())
	}
}