package model

// 用户交易方式
type UserTradingMethod struct {
	ModelBase
	UserID 				uint64 		`gorm:"not null; index" json:"user_id"`				// 用户ID
	TradingMethodID		uint64		`gorm:"not null; index" json:"trading_method_id"`	// 交易方式ID
	BankName			string		`gorm:"size:50" json:"bank_name"`					// 银行名称
	DepositBank			string		`gorm:"size:50" json:"deposit_bank"`				// 开户行名称
	AccountNumber		string		`gorm:"not null" json:"account_number"`				// 账户账号
	Payee				string		`gorm:"not null; size:20" json:"payee"`				// 收款人
	PayeePinyin			string		`gorm:"size:20" json:"payee_pinyin"`				// 收款人拼音
	IsDeleted			bool		`gorm:"default:false" json:"is_deleted"`			// 是否删除
}

// 用户交易方式详细信息
type UserTradingMethodInfo struct {
	ID						string		`json:"id"`							// 用户交易方式ID
	BankName				string		`json:"bank_name"`					// 银行名称
	DepositBank				string		`json:"deposit_bank"`				// 开户行名称
	AccountNumber			string		`json:"account_number"`				// 账户账号
	Payee					string		`json:"payee"`						// 收款人
	PayeePinyin				string		`json:"payee_pinyin"`				// 收款人拼音
	TradingMethodID			uint64 		`json:"-"`							// 交易方式ID
	TradingMethodName 		string		`json:"trading_method_name"`		// 交易方式名称
	TradingMethodEnName  	string 		`json:"trading_method_en_name"`		// 交易人英文名称
	TradingMethodIcon 		string 		`json:"trading_method_icon"`		// 交易方式图标
}