package model

import (
	"fmt"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/spf13/viper"
)

// StatsdClient statsd 客户端
var StatsdClient *statsd.Statter

func init() {
	if viper.GetString("statsd.url") == "" {
		return
	}

	client, err := statsd.NewBufferedClient(viper.GetString("statsd.url"), viper.GetString("statsd.prefix"), 300 * time.Millisecond, 512)
	if err != nil {
		fmt.Println(err.Error())
	}
	StatsdClient = &client
}
