package api

type CommonReq struct {
	Plugin string `json:"plugin"`
	Data interface{} `json:"data"`
}
