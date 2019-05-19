package cron

import (
	"github.com/robfig/cron"
)

var cronMap = map[string]func(){}

func init() {
	cronMap["0 0 0 * * ?"] = everyMorning1ClockCron //每天零晨执行一次
}

// New 构造cron
func New() *cron.Cron {
	c := cron.New()
	for spec, cmd := range cronMap {
		c.AddFunc(spec, cmd)
	}
	return c
}
