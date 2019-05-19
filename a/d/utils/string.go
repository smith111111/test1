package utils

import "strings"

func RemoveSymbols(str string, symbols []string) string {
	for _, symbol := range symbols {
		str = strings.Replace(str, symbol, "", -1)
	}

	return str
}