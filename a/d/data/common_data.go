package data

//服务名定义
const (
	EDGE_SERVER_NAME    = "edge"
	BACKEND_SERVER_NAME = "backend"
	PUSH_SERVER_NAME    = "push"
	IM_SERVER_NAME      = "im"
	TASK_SERVER_NAME    = "task"
	FRONT_SERVER_NAME   = "front"
	STORAGE_SERVER_NAME = "storage"
)

//端口定义
const (
	EDGE_SERVER_PORT    = 8021
	PUSH_SERVER_PORT    = 8024
	BACKEND_SERVER_PORT = 8025
	IM_SERVER_PORT      = 8026
)

var ServerNamePortMap = map[string]int{
	EDGE_SERVER_NAME:    EDGE_SERVER_PORT,
	BACKEND_SERVER_NAME: BACKEND_SERVER_PORT,
	PUSH_SERVER_NAME:    PUSH_SERVER_PORT,
	IM_SERVER_NAME:      IM_SERVER_PORT,
}

type CommonResult struct {
	ErrNo int                    `json:"errNo"`
	Msg   string                 `json:"msg"`
	Data  map[string]interface{} `json:"data,omitempty"`
}
