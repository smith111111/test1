package api

import "encoding/json"

type Request struct {
	Id        string      `json:"id"`
	Appkey    string      `json:"appkey"`
	//Nonce     string      `json:"nonce"`
	NextNonce string      `json:"nextNonce"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	AppData   interface{} `json:"appData"`
}

func (this *Request) Unmarshal(req *CommonReq) error {
	buf, err := json.Marshal(req.Data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buf, this)
	if err != nil {
		return err
	}

	return nil
}
