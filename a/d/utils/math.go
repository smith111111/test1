package utils

import "math"

func IsNaNOrInf(number float64) bool {
	if math.IsNaN(number) || math.IsInf(number, 1) {
		return true
	}
	return false
}