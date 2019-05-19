package api

type Response struct {
	Id     string      `json:"id"`
	Result interface{} `json:"result"`
}
