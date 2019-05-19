package model

type TradingRate struct {
	ModelBase                         // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	TradeType 	int32 		`json:"trade_type"`       									// 交易类型
	TradeMode 	int32 		`json:"trade_mode"` 										// 交易模式
	CoinCode 	string 		`json:"coin_code"` 											// 交易币种
	Rate 		float64 	`gorm:"default: 0"  type:decimal(32,16)" json:"rate"` 		// 手续费率
	MinFee 		float64 	`gorm:"default: 0;type:decimal(32,16)" json:"min_fee"` 		// 最小手续费
	MaxFee 		float64 	`gorm:"default: 0; type:decimal(32,16)" json:"max_fee"` 	// 最大手续费
	FeeCoinCode string 		`json:"fee_coin_code"` 										// 手续费支付币种
	MinAmount 	float64 	`gorm:"default: 0; type:decimal(32,16)" json:"min_amount"` 	// 最小交易额
	MaxAmount 	float64 	`gorm:"default: 0;type:decimal(32,16)"json:"max_amount"`   	// 最大交易额
}

const (
	//购买广告
	TradeTypeBuyAD = 0
	//出售广告
	TradeTypeSaleAD = 1
)

const (
	//购买广告
	TradeTypeBuyADString = "出售广告"
	//出售广告
	TradeTypeSaleADString = "出售广告"
)

func TradeTypeString(typeInt int32) (typeString string) {
	switch typeInt {
	case TradeTypeBuyAD:
		typeString = TradeTypeBuyADString
	case TradeTypeSaleAD:
		typeString = TradeTypeSaleADString
	}
	return
}

const (
	//链上交易
	TradeModeOnChain = 0
	//中心交易
	TradeModeOffChain = 1
)

const (
	//链上交易
	TradeModeOnChainString = "链上交易"
	//中心交易
	TradeModeOffChainString = "中心交易"
)

func TradeModeString(typeInt int32) (typeString string) {
	switch typeInt {
	case TradeModeOnChain:
		typeString = TradeModeOnChainString
	case TradeModeOffChain:
		typeString = TradeModeOffChainString
	}
	return
}

type TradingRateInfo struct {
	ID              uint64    	`json:"id"`                //主键
	TradeType       int32     	`json:"trade_type"`        // 交易类型
	TradeTypeString string  	`json:"trade_type_string"` // 交易类型
	TradeMode       int32     	`json:"trade_mode"`        // 交易模式
	TradeModeString string  	`json:"trade_mode_string"` // 交易模式
	CoinCode        string  	`json:"coin_code"`         // 交易币种
	Rate            float64 	`json:"rate"`              // 手续费率
	MinFee          float64 	`json:"min_fee"`           // 最小手续费
	MaxFee          float64 	`json:"max_fee"`           // 最大手续费
	FeeCoinCode     string  	`json:"fee_coin_code"`     // 手续费支付币种
	MinAmount       float64 	`json:"min_amount"`        // 最小交易额
	MaxAmount       float64 	`json:"max_amount"`        // 最大交易额
}

type TradingRateInfoList []*TradingRateInfo
