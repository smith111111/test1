package log

import (
	"os"
	"flag"
)

func Init() {
	logPath := "./log"
	os.MkdirAll(logPath, os.ModeDir)
	flag.Set("log_dir", logPath)
	flag.Set("log_mode", "stdout:debug,file:debug")
	InitByFlags()
}
