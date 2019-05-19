package log

import (
	"flag"
	"os"
	"testing"
	"time"
)

func init() {
	logPath := "./log"
	os.MkdirAll(logPath, os.ModeDir)
	flag.Set("logf_dir", logPath)
	flag.Set("logmode", "stdout:debug,file:debug")
	InitByFlags()
}

func Test_SLog(t *testing.T) {
	for i := 0; i < 10000000; i++ {
		<-time.After(time.Nanosecond)
		Debug("debug msg:", i)
		Debugf("debugf msg %d", i)
		Verbose("Verbose msg:", i)
		Verbosef("Verbosef msg %d", i)
		Info("Info msg:", i)
		Infof("Infof msg %d", i)
		Warn("Warn msg:", i)
		Warnf("Warnf msg %d", i)
		Error("Error msg:", i)
		Errorf("Errorf msg %d", i)
	}

	<-time.After(5 * time.Minute)
}
