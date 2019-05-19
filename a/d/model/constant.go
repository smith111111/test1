package model

// 通用常量
const (
	// MaxOrder 最大的排序号
	MaxOrder = 10000

	// MinOrder 最小的排序号
	MinOrder = 0

	// PageSize 默认每页的条数
	PageSize = 20

	// MaxPageSize 每页最大的条数
	MaxPageSize = 100

	// MinPageSize 每页最小的条数
	MinPageSize = 10
)

const (
	// 每分钟最多能发送的验证码次数
	CaptchaMinuteLimitCount = 1

	// 每个币种能发布的广告次数
	PushOfferLimitCount = 1
)

// redis相关常量, 为了防止从redis中存取数据时key混乱了，在此集中定义常量来作为各key的名字
const (
	// LoginUser 用户信息
	LoginUser = "loginUser"

	// WithdrawDayLimit 用户每天最多能提币的额度
	WithdrawDayLimit = "withdrawDayLimit"

	// CurrencyByCode 根据代币代码获取代币信息
	CurrencyByCode = "currencyByCode"

	// FiatCurrencyByCode 根据法币代码获取法币信息
	FiatCurrencyByCode = "fiatCurrencyByCode"

	// DepositAddress 用户未使用的充值地址
	DepositAddress = "depositAddress"

	// WithdrawTx 提币记录
	WithdrawTx = "withdrawTx"

	// 邮箱验证码
	EmailCaptcha = "emailCaptcha"

	// 手机验证码
	MobileCaptcha = "mobileCaptcha"

	// 每分钟最多能发送的验证码次数
	CaptchaMinuteLimit = "captchaMinuteLimit"

	// 每个币种最多能发布的广告次数
	PushOfferLimit = "pushOfferLimit"
)
