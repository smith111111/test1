package utils

import "runtime"

func GetConfigPath() string {
	path := "/galaxy/config"
	if runtime.GOOS == "windows" {
		path = "./"
	}

	return path
}