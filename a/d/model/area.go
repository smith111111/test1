package model

// area 区域
type Area struct {
	ID        	uint		`gorm:"primary_key" json:"id"`
	Name 		string  	`json:"name"`
	Code 		string 		`json:"code"`
	Status 		int 		`json:"status"`
	Sort 		int 		`gorm:"default: 0" json:"status"`
	PinYin 		string 		`json:"pin_yin"`
	FullPinYin 	string 		`json:"full_pin_yin"`
}


const (

	//启用
	AreaStatusEnable = 0
	//禁用
	AreaStatusDisable = 1
)

const (

	//启用
	AreaStatusEnableString = "启用"
	//禁用
	AreaStatusDisableString = "禁用"
)

func AreaStatusString (status int) (statusString string) {
	switch status {
	case AreaStatusEnable:
		statusString = AreaStatusEnableString
	case AreaStatusDisable:
		statusString = AreaStatusDisableString
	}
	return
}