package cron

import (
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"galaxyotc/gc_services/account_service/controller/deposit"
)

var cronMap = map[string]func(){}

func init() {
	if !viper.GetBool("server.dev") {
		// 凌晨一点执行，分佣发放
		cronMap["0 0 1 * * ?"] = commissionDistributionCron
	} else {
		// FOT TEST: 10分钟执行一次
		cronMap["0 */10 * * * ?"] = commissionDistributionCron
		// 每隔一分钟查询一次链高度更新状态为充值中的记录
		cronMap["0 */1 * * *"] = deposit.UpdatePendingDeposit
	}
}

// New 构造cron
func New() *cron.Cron {
	c := cron.New()
	for spec, cmd := range cronMap {
		c.AddFunc(spec, cmd)
	}
	return c
}
