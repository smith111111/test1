package eospark

const (
	Success            	  	= 0
	BadRequest         		= 400
	Forbidden       		= 403
	TooManyRequest          = 429
	InternalServerError		= 500
	BlockNotExist           = 100001
	TransactionNotExist		= 100002
	AccountNotExist			= 100003
)

var eosParkErrorCodeMap = map[int]string {
	Success:              	"success",
	BadRequest:   			"bad request",
	Forbidden: 				"forbidden",
	TooManyRequest:     	"too many requests",
	BlockNotExist:    		"block not exist",
	TransactionNotExist:    "transaction not exist",
	AccountNotExist:    	"account not exist",
}

func newEosParkErrorMap() map[int]string {
	return eosParkErrorCodeMap
}