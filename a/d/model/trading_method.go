package model

type TradingMethod struct {
	ModelBase // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	Name 			string		`gorm:"not null; size:20" json:"name"`		// 交易方式名称
	EnName  		string 		`gorm:"size:20" json:"en_name"`				// 交易人英文名称
	Icon 			string 		`json:"icon"`								// 交易方式图标
	IsDeleted		bool 		`gorm:"default:false" json:"is_deleted"`	// 是否删除
}
