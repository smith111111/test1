package api

import (
	"encoding/json"
)

type PairReq struct {
	Appkey      string `json:"appkey"`
	Passthrough bool `json:"passthrough"`
	Origin      string `json:"origin"`
}

func (this *PairReq) Unmarshal(req *CommonReq) error {
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