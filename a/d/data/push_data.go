package data

/*
// 列播(单播)
type PushSendMsgReq struct {
	DisplayType string                 `json:"display_type"` //消息类型，notification-通知，message-消息
	Receivers     string               `json:"receivers"`    //接收者id
	AppId       int8                   `json:"appid"`        //应用id
	Title       string                 `json:"title"`        //通知消息的标题
	Text        string                 `json:"text"`         //通知的内容，对应iOS的alter
	Custom      map[string]interface{} `json:"custom"`       //自定义字段（路由）
	ExpireTime  int32                  `json:"expire_time"`  //失效时间
	LoginStatus int8                   `json:"login_status"` //推送对象是否是登录用户-1:未登录，1:登录，9:all,默认为1
}*/

// 列播(单播)
type PushSendMsgReq struct {
	Receivers string
	Text string `json:"text"`
}

// 添加或修改友盟token管理
type PushDeviceInfoReq struct {
	AppId       int32   `json:"appid"`
	DeviceToken string `json:"device_token"`
	PushType    int32   `json:"push_type"`
}

// 注销用户
type PushLogoutReq struct {
	UserId uint64 `json:"userid"` //用户id
	AppId  int32   `json:"appid"`
}

// 推送记录表
type PushPushMsgResp struct {
	MsgId  int64  `json:"msgid"`  // 消息ID
	Title  string `json:"title"`  // 推送的标题
	Text   string `json:"text"`   // 推送的内容
	Custom string `json:"custom"` // 推动自定义数据
}
