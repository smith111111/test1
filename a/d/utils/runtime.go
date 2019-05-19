package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

var (
	src    string
	srcLen int
)

func Caller(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	short := file
	if strings.HasPrefix(file, src) {
		short = file[srcLen:]
	}
	return fmt.Sprintf("%s#L%d", short, line)
}

func FuncName(f interface{}) string {
	name := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), ".")
	return name[len(name)-1]
}

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("utils: init src failed.")
	}
	if index := strings.Index(file, "src"); index > 0 {
		src = file[0 : index+3]
		srcLen = len(src)
	}
}