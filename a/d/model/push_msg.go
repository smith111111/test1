package model

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// 推送记录表
type PushMsg struct {
	MsgId       int64  	`gorm:"column:msgid;type:bigint" json:"msgid"`                   // 消息ID
	MsgType     int32   `gorm:"column:msg_type;type:tinyint(1)" json:"msg_type"`         // 消息类型
	UserId      uint   	`gorm:"column:userid;type:int(11)" json:"userid"`                // 推送的用户ID
	AppId       int32   `gorm:"column:appid;type:tinyint(1)" json:"appid"`               // 应用标识, 1:OTC
	Platform    int32   `gorm:"column:platform;type:tinyint(1)" json:"platform"`         // 所属平台, 1:android, 2:ios
	LoginStatus int32   `gorm:"column:login_status;type:tinyint(1)" json:"login_status"` // 推送登录状态，-1：退出。1登录，9:全部
	Title       string 	`gorm:"column:title;type:varchar(64)" json:"title"`              // 推送的标题
	Text        string 	`gorm:"column:text;type:varchar(100)" json:"text"`               // 推送的内容
	Custom      string 	`gorm:"column:custom;type:text" json:"custom"`                   // 推动自定义数据
	PushMode    int32   `gorm:"column:push_mode;type:tinyint(1)" json:"push_mode"`       // 推送模式, 1:listcast(列播), 2:broadcast(广播),等
	PushStatus  string 	`gorm:"column:push_status;type:varchar(16)" json:"push_status"`  // 推送状态
	ErrorCode   string 	`gorm:"column:error_code;type:varchar(16)" json:"error_code"`    // 推送不成功时错误码
	PushId      string 	`gorm:"column:push_id;type:varchar(32)" json:"push_id"`          // 推送消息唯一标识
	IsDel       int32   `gorm:"column:is_del;type:tinyint(1)" json:"is_del"`             // 是否删除, 0:正常, 1:删除
	InsertTime  int32  	`gorm:"column:insert_time;type:int(11)" json:"insert_time"`      // 插入时间
	UpdateTime  int32  	`gorm:"column:update_time;type:int(11)" json:"update_time"`      // 更新时间
}

var (
	lock    *sync.RWMutex
	nowTick int64
	nowNo   int
)

//修改消息状态
func (p *PushMsg) UpdatePushStatus(msgId int64, pushStatus, pushId, errorCode string) error {
	updateTime := time.Now().Unix()
	if err := DB.Model(&p).Updates(&PushMsg{PushStatus: pushStatus, ErrorCode: errorCode, PushId: pushId, UpdateTime: int32(updateTime)}).Error; err != nil {
		return err
	}

	return nil
}

//得到消息记录id
func GetPushMsgId() int64 {
	lock.Lock()
	defer lock.Unlock()
	nk := time.Now().Unix()

	if nk != nowTick {
		nowTick = nk
		nowNo = 1
	} else {
		nowNo++
	}

	msgId, _ := strconv.Atoi(fmt.Sprintf("%d%d", nowTick, nowNo))

	return int64(msgId)
}

const (
	UMENG_PUSH_URL   = "http://msg.umeng.com/api/send"
	JIGUANG_PUSH_URL = "https://api.jpush.cn/v3/push"
	XIAOMI_PUSH_URL  = "https://api.xmpush.xiaomi.com/v2/message/regid"
)

//列播(单播)
type ListcastPushReq struct {
	DisplayType string                 `json:"display_type"` //消息类型，notification-通知，message-消息
	UserId      uint                   `json:"userid"`       //用户id
	AppId       int32                  `json:"appid"`        //应用id
	Title       string                 `json:"title"`        //通知消息的标题
	Text        string                 `json:"text"`         //通知的内容，对应iOS的alter
	Custom      map[string]interface{} `json:"custom"`       //自定义字段（路由）
	ExpireTime  string                 `json:"expire_time"`  //失效时间
	LoginStatus int32                  `json:"login_status"` //推送对象是否是登录用户-1:未登录，1:登录，9:all,默认为1
}

//保存推送记录
type PushMsgItem struct {
	ListcastPushReq
	MsgId    int64
	Platform int32
	PushMode int32
}

//列播push时需要字段
type PushInfo struct {
	AppKeySecrets map[string]*DeviceTokenAndPackageName //将AppKey&Secret作为map的key,Token组成字符串为value
	Title         string
	Text          string
	DisplayType   string
	ExpireTime    string
	Custom        map[string]interface{}
	AppId         int32
}

type DeviceTokenAndPackageName struct {
	DeviceTokens string
	UserIds      []uint
	MsgIds       []int64
	PackageName  string
}

// 设备类型
type EDEVICE_TYPE = int32

const (
	EDEVICE_TYPE_ANDROID EDEVICE_TYPE = 1
	EDEVICE_TYPE_IOS     EDEVICE_TYPE = 2
	EDEVICE_TYPE_H5      EDEVICE_TYPE = 3
)

// 推送系统类型
type EPUSH_SYS_TYPE = int32

const (
	EDEVICE_SYS_TYPE_UMENG   EPUSH_SYS_TYPE = 0
	EDEVICE_SYS_TYPE_JIGUANG EPUSH_SYS_TYPE = 1
	EDEVICE_SYS_TYPE_XIAOMI  EPUSH_SYS_TYPE = 2
	EDEVICE_SYS_TYPE_GALAXY  EPUSH_SYS_TYPE = 3
)

// 推送消息类型
type EPUSH_MSG_TYPE = int32

const (
	EPUSH_MSG_TYPE_SINGLE    EPUSH_MSG_TYPE = 0
	EPUSH_MSG_TYPE_LISTCAST  EPUSH_MSG_TYPE = 1
	EPUSH_MSG_TYPE_BROADCAST EPUSH_MSG_TYPE = 2
)

const (
	EPUSH_MSG_TYPE_LISTCAST_STR = "listcast"
)

// 推送消息类型
type ELOGIN_STATUS = int32

const (
	ELOGIN_STATUS_NOLOGIN ELOGIN_STATUS = -1
	ELOGIN_STATUS_LOGIN   ELOGIN_STATUS = 1
	ELOGIN_STATUS_ALL     ELOGIN_STATUS = 9
)

const (
	DISPLAY_TYPE_NOTIFICATION = "notification"
	DISPLAY_TYPE_MESSAGE      = "message"
)
