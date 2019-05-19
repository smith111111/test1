package utils

import (
	"strconv"
	"strings"
)

func SplitToUint(s, sep string) []uint {
	vs1 := strings.Split(s, sep)

	vs2 := make([]uint, len(vs1))

	for i, v := range vs1 {
		v1, _ := strconv.Atoi(v)
		vs2[i] = uint(v1)
	}

	return vs2
}

func Uint64sToStrings(vs1 []uint64) []string {
	vs2 := make([]string, len(vs1))

	for i, v := range vs1 {
		vs2[i] = strconv.Itoa(int(v))
	}

	return vs2
}