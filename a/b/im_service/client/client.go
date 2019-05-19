package client

import (
	"github.com/gao88/netease-im"
	"github.com/spf13/viper"
	searchService "galaxyotc/common/service/search_service"
)

var (
	ImClient *netease.ImClient
	SearchServiceClient *searchService.Client
)

func Init() {
	ImClient = netease.CreateImClient(viper.GetString("im_service.app_key"), viper.GetString("im_service.app_secret"), "")
	SearchServiceClient = searchService.NewClient(viper.GetStringSlice("elastic.urls"))
}