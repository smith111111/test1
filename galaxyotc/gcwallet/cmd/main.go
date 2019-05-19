package main

import (
	"gcwallet"
	"gcwallet/log"
)

func main() {
	log.SetLogPath(".")
	gcw := gcwallet.New()
	gcw.StartScatter()
}

//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"strconv"
//	"time"
//)
//
//type TimeStemp time.Time
//
//
//func (t *TimeStemp) UnmarshalJSON(data []byte) (err error) {
//	unix, err := strconv.ParseInt(string(data), 10, 64)
//	tm:=time.Unix(unix, 0)
//	*t = TimeStemp(tm)
//	return
//}
//
//func (t TimeStemp) MarshalJSON() ([]byte, error) {
//	unix:=strconv.FormatInt(time.Time(t).Unix(),10)
//	return []byte(unix), nil
//}
//
//func (t TimeStemp) String() string {
//	return strconv.FormatInt(time.Time(t).Unix(),10)
//}
//
//type Txn struct {
//	// Transaction ID
//	Txid string
//
//	// The value relevant to the wallet
//	Value string // ZJ: int64
//
//	// The height at which it was mined
//	Height int32
//
//	// The time the transaction was first seen
//	Timestamp TimeStemp
//
//	// This transaction only involves a watch only address
//	WatchOnly bool
//
//	// The number of confirmations on a transaction. This does not need to be saved in
//	// the database but should be calculated when the Transactions() method is called.
//	Confirmations int64
//
//	// The state of the transaction (confirmed, unconfirmed, dead, etc). Implementations
//	// have some flexibility in describing their transactions. Like confirmations, this
//	// is best calculated when the Transactions() method is called.
//	Status string
//
//	// If the Status is Error the ErrorMessage should describe the problem
//	ErrorMessage string
//
//	// Raw transaction bytes
//	Bytes []byte
//}
//
//func main(){
//	//gcw:=gcwallet.New()
//	//if gcw.HasConfig(".") {
//	//	fmt.Println("has config")
//	//}
//	//
//	////msg:=gcw.CreateConfig(".", "soup arch join universe table nasty fiber solve hotel luggage double clean tell oppose hurry weather isolate decline quick dune song enforce curious menu","password")
//	//// msg:=gcw.CreateConfig(".", "wolf dragon lion stage rose snow sand snake kingdom hand daring flower foot walk sword","password")
//	//msg:=gcw.CreateConfig(".", "label pyramid flat spike course crystal humor throw rug frozen food comic","password")
//	//fmt.Println(msg)
//	//msg=gcw.LoadConfig(".", "password")
//	//fmt.Println(msg)
//	//
//	//
//	//gcw.Start()
//	////result :=gcw.GcBalance()
//	//// resultStr = [self.wallet ethSpend:@"10000000000000000" addr:@"0xA07b0902F389854487ac523f4ce64137064347A2" feeLevel:@"economic"];
//	//result:=gcw.GcSpend("9000000000000000000", "0xC5eaAE640FF1acDD846aeF2702C1180e8D2D8fbE","economic")
//	//fmt.Println(result)
//	//gcw.Close()
//
//
//	//iban:=util.NewIbanFromAddress("0x00c5496aee77c1ba1f0854206a26dda82a81d6d8")
//	//fmt.Println(iban)
//
//	now:=time.Now()
//	fmt.Println(now)
//	txn:=&Txn{
//		Timestamp: TimeStemp(now),
//	}
//
//	data, _:=json.Marshal(txn)
//	fmt.Println(string(data))
//
//	var txn2 Txn
//	json.Unmarshal(data, &txn2)
//	fmt.Println(txn2)
//}
