package util

//import (
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/martinlindhe/base36"
//	"regexp"
//
//	"strconv"
//	"strings"
//)
//type Iban struct {
//	iban string `json:"iban"`
//}
//
///**
//* Prepare an IBAN for mod 97 computation by moving the first 4 chars to the end and transforming the letters to
//* numbers (A = 10, B = 11, ..., Z = 35), as specified in ISO13616.
//*
//* @method iso13616Prepare
//* @param {String} iban the IBAN
//* @returns {String} the prepared IBAN
//*/
//func iso13616Prepare(iban string) string {
//	A := 'A'
//	Z := 'Z'
//
//	iban = strings.ToUpper(iban)
//	// iban = iban.substr(4) + iban.substr(0,4);
//	iban = iban[4:] + iban[:4]
//
//	var result []rune
//	for _, runeValue := range iban {
//		code := runeValue;
//		if code >= A && code <= Z{
//			// A = 10, B = 11, ... Z = 35
//			result = append(result, code - A + 10)
//		} else {
//			result = append(result, code)
//		}
//	}
//	return string(result)
//}
//
///**
//* Calculates the MOD 97 10 of the passed IBAN as specified in ISO7064.
//*
//* @method mod9710
//* @param {String} iban
//* @returns {Number}
//*/
//func mod9710(iban string) int64 {
//	remainder := iban
//
//	for len(remainder) > 2 {
//		block:=remainder[:9]
//		blockInt64, _:= strconv.ParseInt(block, 10, 64)
//		remainder = strconv.FormatInt(blockInt64 % 97, 10) + remainder[len(block):]
//	}
//
//	rInt64, _:=strconv.ParseInt(remainder, 10, 64)
//	return rInt64 % 97
//}
//
///**
//* This prototype should be used to create iban object from iban correct string
//*
//* @param {String} iban
//*/
//func NewIban(iban string)  *Iban{
//	return &Iban{iban:iban}
//}
//
///**
//* This method should be used to create iban object from ethereum address
//*
//* @method fromAddress
//* @param {String} address
//* @return {Iban} the IBAN object
//*/
//// web3.eth.Iban.fromEthereumAddress('0x00c5496aee77c1ba1f0854206a26dda82a81d6d8')
//func NewIbanFromAddress(address string)  *Iban {
//
//	asBn, _ := strconv.ParseInt(address[2:], 16, 64)
//	base36 := strconv.FormatInt(asBn, 36)
//	padded := common.LeftPadBytes([]byte(base36), 15)
//	return NewIbanFromBban(strings.ToUpper(string(padded)))
//
//	//bytes, err:=hex.DecodeString(address[2:])
//	//if err !=nil{
//	//	str:=err.Error()
//	//	fmt.Println(str)
//	//}
//	//base36Bytes:=base36.EncodeBytes(bytes)
//	//padded := common.LeftPadBytes([]byte(base36Bytes), 15)
//	//return NewIbanFromBban(strings.ToUpper(string(padded)))
//}
//
///**
//* Convert the passed BBAN to an IBAN for this country specification.
//* Please note that <i>"generation of the IBAN shall be the exclusive responsibility of the bank/branch servicing the account"</i>.
//* This method implements the preferred algorithm described in http://en.wikipedia.org/wiki/International_Bank_Account_Number#Generating_IBAN_check_digits
//*
//* @method fromBban
//* @param {String} bban the BBAN to convert to IBAN
//* @returns {Iban} the IBAN object
//*/
//func NewIbanFromBban(bban string) *Iban {
//	countryCode := "XE"
//	remainder := mod9710(iso13616Prepare(countryCode + "00" + bban))
//	checkDigitStr:="0" + strconv.FormatInt(98 - remainder, 10)
//	var checkDigit = checkDigitStr[len(checkDigitStr)-2:]
//
//	return NewIban(countryCode + checkDigit + bban)
//}
//
///**
//* Should be used to create IBAN object for given institution and identifier
//*
//* @method createIndirect
//* @param {Object} options, required options are "institution" and "identifier"
//* @return {Iban} the IBAN object
//*/
//func NewIbanCreateIndirect(institution,identifier string) *Iban{
//	return NewIbanFromBban("ETH" + institution + identifier)
//}
//
///**
//* Thos method should be used to check if given string is valid iban object
//*
//* @method isValid
//* @param {String} iban string
//* @return {Boolean} true if it is valid IBAN
//*/
//func IbanIsValid(iban string) bool {
//	var i = NewIban(iban)
//	return i.IsValid()
//}
//
///**
//* Should be called to check if iban is correct
//*
//* @method isValid
//* @returns {Boolean} true if it is, otherwise false
//*/
//func (iban * Iban) IsValid() bool {
//	matched, _ := regexp.MatchString("^XE[0-9]{2}(ETH[0-9A-Z]{13}|[0-9A-Z]{30,31})$", iban.iban)
//	return matched
//}
//
///**
//* Should be called to check if iban number is direct
//*
//* @method isDirect
//* @returns {Boolean} true if it is, otherwise false
//*/
//func (iban *Iban) isDirect() bool {
//	return len(iban.iban) == 34 || len(iban.iban) == 35
//}
//
///**
//* Should be called to check if iban number if indirect
//*
//* @method isIndirect
//* @returns {Boolean} true if it is, otherwise false
//*/
//func (iban *Iban) isIndirect() bool {
//	return len(iban.iban) == 20
//}
//
///**
//* Should be called to get iban checksum
//* Uses the mod-97-10 checksumming protocol (ISO/IEC 7064:2003)
//*
//* @method checksum
//* @returns {String} checksum
//*/
//func (iban *Iban) checksum () string {
//	return iban.iban[2:2+2]
//}
//
///**
//* Should be called to get institution identifier
//* eg. XREG
//*
//* @method institution
//* @returns {String} institution identifier
//*/
//func (iban *Iban) institution() string {
//	// return this.isIndirect() ? this._iban.substr(7, 4) : '';
//	if iban.isIndirect() {
//		return iban.iban[7:7+4]
//	}
//	return ""
//}
//
///**
//* Should be called to get client identifier within institution
//* eg. GAVOFYORK
//*
//* @method client
//* @returns {String} client identifier
//*/
//func (iban *Iban) client() string {
//	if iban.isIndirect() {
//		return iban.iban[11:]
//	}
//	return ""
//};
//
///**
//* Should be called to get client direct address
//*
//* @method address
//* @returns {String} client direct address
//*/
//func (iban *Iban) address() string {
//	if iban.isDirect() {
//		//var base36 = this._iban.substr(4);
//		//var asBn = new BigNumber(base36, 36);
//		//return padLeft(asBn.toString(16), 20);
//
//		var base36Number = iban.iban[4:]
//		var asBn = base36.Decode(base36Number)
//		result:= common.LeftPadBytes([]byte(strconv.FormatInt(int64(asBn), 16)), 15)
//		return string(result)
//	}
//
//	return ""
//};
//
//func (iban *Iban) String() string {
//	return iban.iban
//}