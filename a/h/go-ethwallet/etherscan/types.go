package etherscan

import (
	"encoding/json"
	"fmt"
	"time"
	"strconv"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

// Hash

type Hash struct {
	common.Hash
}

func (h *Hash) UnmarshalJSON(data []byte) (err error) {
	h.Hash = common.HexToHash(strings.Trim(string(data), `"`))
	return nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return h.Hash.Bytes(), nil
}

// Address

type Address struct {
	common.Address
}

func (a *Address) UnmarshalJSON(data []byte) (err error) {
	a.Address = common.HexToAddress(strings.Trim(string(data), `"`))
	return nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	return a.Address.Bytes(), nil
}

// JSONTime

type JSONTime struct {
	time.Time
}

const JSONTimeFormat = "2006-01-02T15:04:05"

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	timestamp, err := strconv.ParseInt(strings.Trim(string(data), `"`), 10, 64)
	if err != nil {
		return err
	}

	t.Time = time.Unix(timestamp, 0)
	return nil
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Format(JSONTimeFormat))), nil
}

type TransactionStatus uint8

const (
	TransactionStatusPending TransactionStatus = iota
	TransactionStatusSuccess
	TransactionStatusFail
	TransactionStatusUnknown = TransactionStatus(255)
)

func (s *TransactionStatus) UnmarshalJSON(data []byte) error {
	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	switch decoded {
	case "0":
		*s = TransactionStatusPending
	case "1":
		*s = TransactionStatusSuccess
	case "2":
		*s = TransactionStatusFail
	default:
		*s = TransactionStatusUnknown
	}
	return nil
}

func (s TransactionStatus) MarshalJSON() (data []byte, err error) {
	out := "unknown"
	switch s {
	case TransactionStatusPending:
		out = "pending"
	case TransactionStatusSuccess:
		out = "success"
	case TransactionStatusFail:
		out = "fail"
	}
	return json.Marshal(out)
}
func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "pending"
	case TransactionStatusSuccess:
		return "success"
	case TransactionStatusFail:
		return "fail"
	default:
		return "unknown"
	}
}