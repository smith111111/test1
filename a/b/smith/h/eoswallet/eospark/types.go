package eospark

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// For reference:
// https://github.com/mithrilcoin-io/EosCommander/blob/master/app/src/main/java/io/mithrilcoin/eoscommander/data/remote/model/types/EosByteWriter.java

type Name string
type AccountName Name
type PermissionName Name
type ActionName Name
type TableName Name
type ScopeName Name

func AN(in string) AccountName    { return AccountName(in) }
func ActN(in string) ActionName   { return ActionName(in) }
func PN(in string) PermissionName { return PermissionName(in) }

type AccountResourceLimit struct {
	Used      uint64 `json:"used"`
	Available uint64 `json:"available"`
	Max       uint64 `json:"max"`
}

type AccountStaked struct {
	NetWeight      	string 		`json:"net_weight"`
	Available 		string 		`json:"available"`
}

type AccountUnStaked struct {
	NetAmount      	string 		`json:"net_amount"`
	CpuAmount 		string 		`json:"cpu_amount"`
}

type PermissionLevel struct {
	Actor      AccountName    `json:"actor"`
	Permission PermissionName `json:"permission"`
}

// NewPermissionLevel parses strings like `account@active`,
// `otheraccount@owner` and builds a PermissionLevel struct. It
// validates that there is a single optional @ (where permission
// defaults to 'active'), and validates length of account and
// permission names.
func NewPermissionLevel(in string) (out PermissionLevel, err error) {
	parts := strings.Split(in, "@")
	if len(parts) > 2 {
		return out, fmt.Errorf("permission %q invalid, use account[@permission]", in)
	}

	if len(parts[0]) > 12 {
		return out, fmt.Errorf("account name %q too long", parts[0])
	}

	out.Actor = AccountName(parts[0])
	out.Permission = PermissionName("active")
	if len(parts) == 2 {
		if len(parts[1]) > 12 {
			return out, fmt.Errorf("permission %q name too long", parts[1])
		}

		out.Permission = PermissionName(parts[1])
	}

	return
}

// JSONTime

type JSONTime struct {
	time.Time
}

const JSONTimeFormat = "2006-01-02T15:04:05"

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Format(JSONTimeFormat))), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	if string(data) == "null" {
		return nil
	}

	t.Time, err = time.Parse(`"`+JSONTimeFormat+`"`, string(data))
	return err
}

// ParseJSONTime will parse a string into a JSONTime object
func ParseJSONTime(date string) (JSONTime, error) {
	var t JSONTime
	var err error
	t.Time, err = time.Parse(JSONTimeFormat, string(date))
	return t, err
}

// HexBytes

type HexBytes []byte

func (t HexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(t))
}

func (t *HexBytes) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	*t, err = hex.DecodeString(s)
	return
}

func (t HexBytes) String() string {
	return hex.EncodeToString(t)
}

// Checksum256

type Checksum160 []byte

func (t Checksum160) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(t))
}
func (t *Checksum160) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	*t, err = hex.DecodeString(s)
	return
}

type Checksum256 []byte

func (t Checksum256) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(t))
}
func (t *Checksum256) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	*t, err = hex.DecodeString(s)
	return
}

func (t Checksum256) String() string {
	return hex.EncodeToString(t)
}

type Checksum512 []byte

func (t Checksum512) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(t))
}
func (t *Checksum512) UnmarshalJSON(data []byte) (err error) {
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return
	}

	*t, err = hex.DecodeString(s)
	return
}

type TransactionStatus uint8

const (
	TransactionStatusExecuted TransactionStatus = iota ///< succeed, no error handler executed
	TransactionStatusSoftFail                          ///< objectively failed (not executed), error handler executed
	TransactionStatusHardFail                          ///< objectively failed and error handler objectively failed thus no state change
	TransactionStatusDelayed                           ///< transaction delayed
	TransactionStatusExpired                           ///< transaction expired
	TransactionStatusUnknown  = TransactionStatus(255)
)

func (s *TransactionStatus) UnmarshalJSON(data []byte) error {
	var decoded string
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	switch decoded {
	case "executed":
		*s = TransactionStatusExecuted
	case "soft_fail":
		*s = TransactionStatusSoftFail
	case "hard_fail":
		*s = TransactionStatusHardFail
	case "delayed":
		*s = TransactionStatusDelayed
	case "expired":
		*s = TransactionStatusExpired
	default:
		*s = TransactionStatusUnknown
	}
	return nil
}

func (s TransactionStatus) MarshalJSON() (data []byte, err error) {
	out := "unknown"
	switch s {
	case TransactionStatusExecuted:
		out = "executed"
	case TransactionStatusSoftFail:
		out = "soft_fail"
	case TransactionStatusHardFail:
		out = "hard_fail"
	case TransactionStatusDelayed:
		out = "delayed"
	case TransactionStatusExpired:
		out = "expired"
	}
	return json.Marshal(out)
}
func (s TransactionStatus) String() string {

	switch s {
	case TransactionStatusExecuted:
		return "executed"
	case TransactionStatusSoftFail:
		return "soft_fail"
	case TransactionStatusHardFail:
		return "hard_fail"
	case TransactionStatusDelayed:
		return "delayed"
	case TransactionStatusExpired:
		return "expired"
	default:
		return "unknown"
	}

}