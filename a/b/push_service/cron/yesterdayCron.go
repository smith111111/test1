package cron

import (
	"time"
	"galaxyotc/common/model"
	"galaxyotc/common/log"
)

// 每天零晨执行一次
func everyMorning1ClockCron() {
	log.Debug("everyMorning1ClockCron...")

	// 删除3个月前的历史消息
	insertTime := time.Now().AddDate(0, -3, 0).Unix()
	if err := model.DB.Delete(&model.PushMsg{}).Where("insert_time < ?", insertTime).Error; err != nil {
		log.Error(err)
	}
}
