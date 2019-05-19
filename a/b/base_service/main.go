package main

import (
	_ "galaxyotc/gc_services/base_service/init"

	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"

	"galaxyotc/gc_services/base_service/router"

	"galaxyotc/common/log"

)

func main() {
	log.Infof("galaxy base service is running")

	// Creates a router without any middleware by default
	app := gin.New()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	maxSize := viper.GetInt64("server.max_multipart_memory")
	app.MaxMultipartMemory = maxSize << 20 // 3 MiB

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	app.Use(gin.Recovery())

	app.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Length", "Content-Type", "AppKey"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Route(app)

	app.Run(":" + fmt.Sprintf("%d", viper.GetInt("base_service.port")))
}
