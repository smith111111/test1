package data

// 注册聊天用户
type ImUserRegisterReq struct {
	Id     uint64                 `json:"id"`     //用户ID
	Name   string                 `json:"name"`   //用户名
	Props  map[string]interface{} `json:"props"`  //json属性
	Icon   string                 `json:"icon"`   //头像
	Email  string                 `json:"email"`  //邮箱
	Birth  string                 `json:"birth"`  //生日
	Mobile string                 `json:"mobile"` //手机号
	Gender int8                   `json:"gender"` //年龄
	Ex     map[string]interface{} `json:"ex"`     //用户名片扩展字段
}

// 批量发送自定义系统消息
type ImSendSysMsgReq struct {
	From   uint64                 `json:"from"`   //发送者
	To     uint64                 `json:"to"`     //接收者
	Attach map[string]interface{} `json:"attach"` //自定义通知内容，最大总数据长度4096字符
}

// IM历史信息
type ImHistoryMsgResp struct {
	MsgId        int64  `json:"msgid"`         // 消息ID
	MsgType      string `json:"msg_type"`      // 消息类型
	Body         string `json:"body"`          // 消息内容
	Attach       string `json:"attach"`        // 附加消息
	Ext          string `json:"ext"`           // 消息扩展字段
	MsgTimestamp int64 `json:"msgTimestamp "` // 消息发送时间
}

// 抄送消息
type UserRegisterResp struct {
	Token string `json:"token"`
}

// 抄送消息
type MsgCopyInfoResp struct {
	EventType      string `json:"eventType"`      //抄送消息类型
	ConvType       string `json:"convType"`       //会话具体类型
	To             string `json:"to"`             //消息接收者的用户账号
	FromAccount    string `json:"fromAccount"`    //消息发送者的用户账号
	FromClientType string `json:"fromClientType"` //发送客户端类型
	FromDeviceId   string `json:"fromDeviceId"`   //发送设备id
	FromNick       string `json:"fromNick"`       //发送方昵称
	MsgTimestamp   string `json:"msgTimestamp"`   //消息发送时间
	MsgType        string `json:"msgType"`        //会话具体通知消息类型
	Body           string `json:"body"`           //消息内容
	Attach         string `json:"attach"`         //附加消息
	MsgIdClient    string `json:"msgidClient"`    //客户端生成的消息id
	MsgIdServer    string `json:"msgidServer"`    //服务端生成的消息id
	ResendFlag     string `json:"resendFlag"`     //重发标记
	CustomApnsText string `json:"customApnsText"` //自定义系统通知消息推送文本
	Ext            string `json:"ext"`            //消息扩展字段
}
