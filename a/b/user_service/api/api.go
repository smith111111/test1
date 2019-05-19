package api

import (
	//imApi "galaxyotc/common/service/im_service"
	//pushApi "galaxyotc/common/service/push_service"
	privateWalletApi "galaxyotc/common/service/wallet_service/private_wallet_service"
	"github.com/nats-io/go-nats"
)

var (
	//ImApi *imApi.Client
	//PushApi *pushApi.Client
	PrivateWalletApi *privateWalletApi.Client
)

func Init(nc *nats.Conn) {
	//ImApi = imApi.NewClient(nc)
	//PushApi = pushApi.NewClient(nc)
	PrivateWalletApi = privateWalletApi.NewClient(nc)
}