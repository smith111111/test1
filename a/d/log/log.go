/*
slog is a micro log libray.log format is use default.
log use ConsolePrinter as default(least level) ,you can use it without any configuration.

in advance,use AddPrinter config you output mode.once operated the default
console printer is no longer exist.so add it if needed.

	-logmode=stdout:info,file:warn
	-logf_dir=.
	-logf_name=app
	-logf_ksize=10
	-logf_blockmillis=200
	-logf_bufferrow=123

Attention: when init log by flags,you must call InitByFlags.(since we don't known when flags is ready).
And slog's file output use one goroutine,it not worth to support flush file by hand which means a lock is needed.
if you want guarantee all log flush to file before exit,I suggest just wait for a few second .
*/
package log

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type flags struct {
	out              string
	file_dir         string
	file_name        string
	file_ksize       int
	file_blockmillis int
	file_bufferrow   int
	file_backup      int
}

func (this *flags) parse() []*levPrinter {
	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(os.Stderr, "parse slog flag failed: %v", e)
			os.Exit(0)
		}
	}()

	this.out = strings.ToUpper(this.out)
	am := strings.Split(this.out, ",")
	l := make([]*levPrinter, 0)
	for _, m := range am {
		ml := strings.Split(m, ":")
		if ml[0] == "FILE" {
			p, _ := _NewFilePrinter(this.file_bufferrow, this.file_dir, this.file_name,
				this.file_ksize, this.file_backup, this.file_blockmillis)
			l = append(l, &levPrinter{stringLev(ml[1]), p})
		} else if ml[0] == "STDOUT" {
			l = append(l, &levPrinter{stringLev(ml[1]), &_ConsolePrinter{}})
		}
	}
	return l
}

var _flags flags

/*
calling when configurate by flags
*/
func InitByFlags() {
	if !flag.Parsed() {
		flag.Parse()
	}
	alp := _flags.parse()
	for _, lp := range alp {
		_AddPrinter(lp.l, lp.p)
	}
}

type lev int

const (
	LevDEBUG   lev = 1
	LevVERBOSE lev = 2
	LevINFO    lev = 4
	LevWARN    lev = 8
	LevERROR   lev = 0x10
	LevFATAL   lev = 0x11
)

func (this *lev) String() string {
	switch *this {
	case LevDEBUG:
		return "DEBUG"
	case LevVERBOSE:
		return "VERBO"
	case LevINFO:
		return "INFO"
	case LevWARN:
		return "WARN"
	case LevERROR:
		return "ERROR"
	default:
		return "FATAL"
	}
}

func stringLev(l string) lev {
	l = strings.ToUpper(l)
	switch l {
	case "DEBUG":
		return LevDEBUG
	case "ERROR":
		return LevERROR
	case "INFO":
		return LevINFO
	case "WARN":
		return LevWARN
	case "FATAL":
		return LevFATAL
	default:
		return LevVERBOSE
	}
}

var (
	rootLev lev
	errch   chan *string = make(chan *string)
	f       _Formater    = new(_DefaultFormater)
	lp      []*levPrinter
	mux     sync.RWMutex
	config  bool
)

func init() {
	flag.StringVar(&_flags.out, "logmode", "stdout:debug,file:warn", "out mode,like this: stdout:info,file:warn ")
	flag.StringVar(&_flags.file_dir, "logf_dir", "./log", "log file dir")
	flag.StringVar(&_flags.file_name, "logf_name", "", "log file name")
	flag.IntVar(&_flags.file_ksize, "logf_ksize", 4*1024, "log file max size ,kB unit")
	flag.IntVar(&_flags.file_blockmillis, "logf_blockmillis", 10, "file loging blocked working thread if output is busy")
	flag.IntVar(&_flags.file_bufferrow, "logf_bufferrow", (1024 * 256), "file logger cached row number without blocking if output is busy")
	flag.IntVar(&_flags.file_backup, "logf_backup", 1000, "log file max backup count")

	lp = []*levPrinter{
		&levPrinter{LevDEBUG, &_ConsolePrinter{}},
	}

	rootLev = LevDEBUG

	go flusherr()
}

func _AddPrinter(l lev, p _Printer) {
	mux.Lock()
	defer mux.Unlock()
	if !config {
		config = true
		lp = make([]*levPrinter, 0)
		rootLev = LevERROR
	}
	lp = append(lp, &levPrinter{l, p})
	if l < rootLev {
		rootLev = l
	}
}

func Debug(args ...interface{}) {
	out(LevDEBUG, args...)
}
func Verbose(args ...interface{}) {
	out(LevVERBOSE, args...)
}
func Info(args ...interface{}) {
	out(LevINFO, args...)
}
func Warn(args ...interface{}) {
	out(LevWARN, args...)
}
func Error(args ...interface{}) {
	out(LevERROR, args...)
}

func Fatal(args ...interface{}) {
	out(LevFATAL, args...)
	printStack(true)
	<-time.After(2 * time.Second)

	os.Exit(-1)
}

func Fatalf(format string, args ...interface{}) {
	outf(LevFATAL, format, args...)
	printStack(true)
	<-time.After(2 * time.Second)
	os.Exit(-1)
}

func Debugf(format string, args ...interface{}) {
	outf(LevDEBUG, format, args...)
}
func Verbosef(format string, args ...interface{}) {
	outf(LevVERBOSE, format, args...)
}
func Infof(format string, args ...interface{}) {
	outf(LevINFO, format, args...)
}
func Warnf(format string, args ...interface{}) {
	outf(LevWARN, format, args...)
}
func Errorf(format string, args ...interface{}) {
	outf(LevERROR, format, args...)
}

func out(lv lev, args ...interface{}) {
	if rootLev <= lv {
		s := f.Format(lv, args...)
		mux.RLock()
		defer mux.RUnlock()
		for _, n := range lp {
			if lv >= n.l {
				n.p.Print(s)
			}
		}
	}
}

func outf(lv lev, format string, args ...interface{}) {
	if rootLev <= lv {
		s := f.Formatf(lv, format, args...)
		mux.RLock()
		defer mux.RUnlock()
		for _, n := range lp {
			if lv >= n.l {
				n.p.Print(s)
			}
		}
	}
}

func printStack(all bool) {
	n := 5000
	if all {
		n = 50000
	}
	var trace []byte

	for i := 0; i < 5; i++ {
		n *= 2
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes <= len(trace) {
			n = nbytes
			break
		}
	}
	ms := string(trace[:n])
	mux.RLock()
	defer mux.RUnlock()
	for _, n := range lp {
		n.p.Print(&ms)
	}
}

type levPrinter struct {
	l lev
	p _Printer
}

//when log system has error,we will try record as much as possible(unrealize)
func slogerr(s *string) {
	select {
	case errch <- s:
	default:
		fmt.Fprint(os.Stderr, "slogerr timeout:"+*s)
	}
}
func flusherr() {
	for s := range errch {
		fmt.Fprint(os.Stderr, *s)
	}
}
