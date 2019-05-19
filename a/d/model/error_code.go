package model

type errorCode struct {
	SUCCESS      int
	ERROR        int
	NotFound     int
	LoginError   int
	LoginTimeout int
}

// ErrorCode 错误码
var ErrorCode = errorCode{
	SUCCESS:      0,
	ERROR:        1,
	NotFound:     404,
	LoginTimeout: 1001, //登录超时
}
