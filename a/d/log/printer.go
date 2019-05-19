package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type _Printer interface {
	Print(s *string)
}

type _ConsolePrinter struct {
}

func (this *_ConsolePrinter) Print(s *string) {
	fmt.Fprint(os.Stderr, *s)
}

type _UdpNetPrinter struct {
}

func (this *_UdpNetPrinter) Print(s *string) {
	fmt.Fprint(os.Stderr, *s)
}

type _FilePrinter struct {
	ch          chan *string
	blockMillis time.Duration
	baseName    string
	file        *os.File
	rsize       int
	csize       int
	backup      int
}

/*
文件输出器
参数：
	maxrn:缓冲队列长度(记录条数)
	dir:日志输出目录,""为当前
	name:日志文件名称,""为程序名
	ksize:日志文件滚动大小，以KB为单位
	backup:日志文件滚动个数
	blockMillis:缓冲队列满时，日志线程阻塞工作线程最大毫秒数。越大则丢日志的可能性越低
*/
func _NewFilePrinter(maxrn int, dir string, name string, ksize int, backup int, blockMillis int) (_Printer, error) {
	var e error
	if name == "" {
		name = filepath.Base(os.Args[0])
		if i := strings.LastIndex(name, "."); i != -1 {
			name = name[:i]
		}
	}
	if blockMillis < 10 {
		blockMillis = 10
	}
	if backup < 1 {
		backup = 1
	}
	p := &_FilePrinter{
		ch:          make(chan *string, maxrn),
		rsize:       ksize * 1024,
		blockMillis: time.Millisecond * time.Duration(blockMillis),
		backup:      backup,
	}
	dir, e = filepath.Abs(dir)
	if e != nil {
		return nil, e
	}
	e = os.MkdirAll(dir, os.ModeDir)
	if e != nil {
		return nil, e
	}
	p.baseName = filepath.Clean(dir + string(filepath.Separator) + name + ".log")
	p.file, e = os.OpenFile(p.baseName, os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		return nil, e
	}
	s := "--------------------start--------------------\n"
	p.Print(&s)
	go p.flush()
	return p, nil
}

func (this *_FilePrinter) Print(s *string) {
	//fmt.Println(">>>>>>>>>>", *s)
	select {
	case this.ch <- s:
	case <-time.After(this.blockMillis):
		es := "FilePrinter enqueue timeout:" + *s
		slogerr(&es)
	}
}

func (this *_FilePrinter) checkfile() {
	if this.csize > this.rsize {
		this.file.Sync()
		this.file.Close()
		this.roll()
		this.file, _ = os.OpenFile(this.baseName, os.O_CREATE|os.O_WRONLY, 0666)
		this.csize = 0
	}
}

//func (this *_FilePrinter) rename(t *time.Time) {
//	_, m, d := t.Date()
//	ts := fmt.Sprintf(".%02d%02d-%02d%02d%02d%03d", m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000)
//	os.Rename(this.baseName, this.baseName+ts)
//}

func (this *_FilePrinter) roll() {
	if this.backup == 0 {
		os.Remove(this.baseName)
		return
	}
	os.Remove(this.baseName + "." + strconv.Itoa(this.backup))
	for i := this.backup - 1; i > 0; i-- {
		o := this.baseName + "." + strconv.Itoa(i)
		n := this.baseName + "." + strconv.Itoa(i+1)
		os.Rename(o, n)
	}
	os.Rename(this.baseName, this.baseName+".1")
}

func (this *_FilePrinter) flush() {
	var n int
	var e error
	for {
		n = 0
	loop:
		for {
			select {
			case s := <-this.ch:
				this.checkfile()
				n, e = this.file.WriteString(*s)
				if e != nil {
					this.csize = this.rsize + 1
					break
				} else {
					this.csize += n
				}
			default:
				break loop
			}
		}
		if n != 0 {
			this.file.Sync()
		}
		<-time.After(time.Second)
	}
}
