package model

import "time"

type RealnameVerification struct {
	ModelBase                                                    // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	AreaCode     string     	`json:"area_code"`                   // 用户名
	IDCardType   int32       	`json:"id_card_type"`                // 身份照或护照
	IDCardNo     string     	`gorm:"size:50" json:"id_card_no"`   // 身份证号
	IDCardName   string     	`gorm:"size:50" json:"id_card_name"` //  身份证人名
	Status       int32        	`json:"status"`                      // 状态
	ApplyTime    time.Time  	`json:"apply_time"`                  // 申请时间
	ApprovedTime *time.Time 	`json:"approved_time"`               // 审批时间
	ApprovedUser string     	`json:"approved_user"`               //审批人
	Remark  	 string  		`json:"remark"`						 //审批备注
	UserId       uint64       	`json:"user_id"`                     // 用户ID
}

type ApproveRequestModel struct {
	ID 				uint    		`json:"id"`
	Status 			int32 			`json:"status"`
	Remark  	 	string  		`json:"remark"`
	ApprovedUser 	string			`json:"approved_user"`
}


// 卡类型
const (
	// 身份证
	CardType_ID = 1
	// 护照
	CardType_Passport = 2
)

const (
	//申请
	Apply = 1
	//审批中
	Approvaling = 2
	//已通过
	Approvaled = 3
	//已拒绝
	Reject = 4
)

const (
	ApplyString       = "申请"
	ApprovalingString = "审批中"
	ApprovaledString  = "已通过"
	RejectString      = "已拒绝"
)
