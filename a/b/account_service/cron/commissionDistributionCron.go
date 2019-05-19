package cron

import (
	"galaxyotc/gc_services/account_service/controller/commission"
	"galaxyotc/common/log"
)

// 佣金发放
func commissionDistributionCron()  {
	// 1, 检查是否还有没生成佣金的订单（订单必须是成功交易的）
	// 2, 发放佣金
 	if _, err := commission.Distribution(); err != nil {
 		log.Errorf("CommissionDistributionCron Error: %s", err.Error())
	}
}
