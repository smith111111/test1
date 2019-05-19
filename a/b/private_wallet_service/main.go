package main

import (
	_ "galaxyotc/gc_services/private_wallet_service/init"

	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"galaxyotc/common/log"
)

func main() {
	log.Infof("galaxy private wallet service is running")

	// Creates a router without any middleware by default
	app := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())

	app.Run(":" + fmt.Sprintf("%d", viper.GetInt("private_wallet_service.port")))
}
