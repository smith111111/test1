package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

var logPath string

func SetLogPath(path string) {
	logPath = path
}

func WriteLog(log string) {
	var filename = logPath + "/gcwallet.log"
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer f.Close()

	head := ""

	pc, _, line, ok := runtime.Caller(1)
	if ok {
		f := runtime.FuncForPC(pc)
		arr1 := strings.Split(f.Name(), ".")
		if len(arr1) > 0 {
			head += arr1[len(arr1)-1]
		} else {
			head += f.Name()
		}

		head += ":"
		head += fmt.Sprintf("%d ", line)
	}

	log = head + " | " + log

	fmt.Println(log)

	log += "\n"

	_, err = io.WriteString(f, log)
	if err != nil {
		return
	}
}
