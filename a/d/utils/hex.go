package utils

import "encoding/hex"

// Bytes2Hex returns the hexadecimal encoding of d.
func BytesToHex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func HexToBytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}