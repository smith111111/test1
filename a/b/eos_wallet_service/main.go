package main

import (
	_ "galaxyotc/gc_services/eos_wallet_service/init"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"galaxyotc/common/log"
)

func main() {
	log.Infof("galaxy eos wallet service is running")

	app := gin.New()
	app.Use(gin.Logger())
	app.Use(gin.Recovery())

	app.Run(":" + fmt.Sprintf("%d", viper.GetInt("eos_wallet_service.port")))
}
