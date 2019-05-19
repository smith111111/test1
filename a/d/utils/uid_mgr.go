package utils

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"sync"
)

var uidMgrService *UidMgrService
var uidMgrOnce sync.Once

func GetUidMgrService() *UidMgrService {
	uidMgrOnce.Do(func() {
		uidMgrService = &UidMgrService{}
		uidMgrService.init()
	})

	return uidMgrService
}

type UidMgrService struct {
	node *snowflake.Node
}

func (this *UidMgrService) init() {
	ip, err := Lower16BitPrivateIP()
	if err != nil {
		return
	}

	this.node, err = snowflake.NewNode(int64(ip))
	if err != nil {
		fmt.Println(err)
	}
}

func (this *UidMgrService) GetNextUserId() uint64 {
	return uint64(this.node.Generate().Int64())
}
