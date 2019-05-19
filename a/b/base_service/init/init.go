package init

import (
	"github.com/spf13/viper"
	"galaxyotc/common/log"
	"galaxyotc/common/model"
	"galaxyotc/common/config"
	"galaxyotc/gc_services/base_service/api"

	commonService "galaxyotc/common/service"
	"galaxyotc/common/utils"
)

func init() {
	// 通用配置
	config.DefaultViper()
	// 当前服务配置
	config.SpecifyViper("base_service", "toml", utils.GetConfigPath())

	// 初始化日志
	log.Init()

	// 初始化数据库
	model.NewDB(viper.GetString("db.gc_dsn"))
	model.NewRedis()

	// 获取nats连接
	natsUrl := viper.GetString("nats.url")
	serviceName := viper.GetString("base_service.name")
	nc := commonService.NewAntsClient(natsUrl, serviceName)

	// 初始化API客户端
	api.Init(nc)
}