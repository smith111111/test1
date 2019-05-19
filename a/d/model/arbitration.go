package model

// 仲裁申请
type Arbitration struct {
	ModelBase // fields `ID`, `CreatedAt`, `UpdatedAt` will be added
	OrderID					uint			`json:"order_id"`				// 订单ID
	OrderSN					string			`json:"order_sn"`				// 订单流水号
	Plaintiff				uint 			`json:"plaintiff"`				// 申诉人ID
	PlaintiffName			string 			`json:"plaintiff_name"`			// 申诉人姓名
	Defendant				uint 			`json:"defendant"`				// 被申诉人ID
	DefendantName			string 			`json:"defendant_name"`			// 被申诉人姓名
	Amount					float64 		`json:"amount"`					// 申诉代币数量
	Currency				uint			`json:"currency"`				// 申诉代币类型
	ArbitrationType			int 			`json:"arbitration_type"`		// 申诉类型
	ArbitrationReason		string 			`json:"arbitration_note"`		// 申诉理由
	Pictures				string 			`json:"pictures"`				// 图片证据
	Arbiter					uint			`json:"arbiter"`				// 仲裁者
	ArbiterName				string 			`json:"arbiter_name"`			// 仲裁者姓名
	ArbitrationResult		uint 			`json:"arbitration_result"`		// 仲裁结果
}

const (
	ArbitrationUndetermined = 0
	ArbitrationSuccessful = 1
	ArbitrationFailure = 2
)

const (
	ArbitrationUndeterminedString = "等待仲裁"
	ArbitrationSuccessfulString = "仲裁成功"
	ArbitrationFailureString = "仲裁失败"
)

func ArbitrationResultString(resultInt int) (resultString string) {
	switch resultInt {
	case ArbitrationUndetermined:
		resultString = ArbitrationUndeterminedString
	case ArbitrationSuccessful:
		resultString = ArbitrationSuccessfulString
	case ArbitrationFailure:
		resultString = ArbitrationFailureString
	}
	return
}