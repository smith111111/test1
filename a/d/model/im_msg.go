package model

// 聊天记录表
type ImMsg struct {
	MsgIdServer    int64  `gorm:"column:msgIdServer" json:"msgIdServer"`       //服务端生成的消息id
	MsgIdClient    string `gorm:"column:msgIdClient" json:"msgIdClient"`       //客户端生成的消息id
	EventType      int8   `gorm:"column:eventType" json:"eventType"`           //抄送消息类型
	ConvType       string `gorm:"column:convType" json:"convType"`             //会话具体类型
	To             string `gorm:"column:to" json:"to"`                         //消息接收者的用户账号
	FromAccount    string `gorm:"column:fromAccount" json:"fromAccount"`       //消息发送者的用户账号
	FromClientType string `gorm:"column:fromClientType" json:"fromClientType"` //发送客户端类型
	FromDeviceId   string `gorm:"column:fromDeviceId" json:"fromDeviceId"`     //发送设备id
	FromNick       string `gorm:"column:fromNick" json:"fromNick"`             //发送方昵称
	MsgTimestamp   int64  `gorm:"column:msgTimestamp" json:"msgTimestamp"`     //消息发送时间
	MsgType        string `gorm:"column:msgType" json:"msgType"`               //会话具体通知消息类型
	Body           string `gorm:"column:body" json:"body"`                     //消息内容
	Attach         string `gorm:"column:attach" json:"attach"`                 //附加消息
	ResendFlag     int8   `gorm:"column:resendFlag" json:"resendFlag"`         //重发标记
	CustomApnsText string `gorm:"column:customApnsText" json:"customApnsText"` //自定义系统通知消息推送文本
	Ext            string `gorm:"column:ext" json:"ext"`                       //消息扩展字段
}