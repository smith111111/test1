package api

import "encoding/json"

type RekeyedReq struct {
	Appkey string `json:"appkey"`
	Origin string `json:"origin"`
}

func (this *RekeyedReq) Unmarshal(req *CommonReq) error {
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