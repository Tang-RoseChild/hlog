package hlog

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"time"
)

func debug_p(a ...interface{}) {
	// fmt.Println(a...)
}

// MsgType const
type MsgType int

const (
	DebugMsg MsgType = iota + 1
	InfoMsg
	ErrorMsg
)

// normal const block
const (
	defaultMax     = 30
	defaultMsgChan = 10
)

type Logger struct {
	w          io.Writer
	debugFlag  bool
	max        int // max length of chan
	msgChan    chan *msg
	cache      []*msg
	cacheIndex int
	done       chan struct{}
	fHead      FormatHeader
	inited     bool          // whether inited
	initedChan chan struct{} // used for setting inited
}

type msg struct {
	c interface{} // content
	t MsgType     // msg type
}

func (t MsgType) String() string {
	var s string
	switch t {
	case DebugMsg:
		s = "debug msg"
	case InfoMsg:
		s = "info msg"
	case ErrorMsg:
		s = "error msg"
	default:
		s = "unknow type"

	}
	return s
}
func New(w io.Writer) *Logger {
	return &Logger{
		w:          w,
		max:        defaultMax,
		msgChan:    make(chan *msg, defaultMsgChan),
		cache:      make([]*msg, defaultMax),
		cacheIndex: 0,
		done:       make(chan struct{}),
		fHead:      defautFormatHeader{},
		initedChan: make(chan struct{}, 1),
	}
}

func (l *Logger) Run() {
	go l.init()
	<-l.initedChan
	l.inited = true
}

func (l *Logger) init() {
	l.initedChan <- struct{}{}
	for {
		select {
		case msg := <-l.msgChan:
			debug_p("in l.msgchan")
			l.cache[l.cacheIndex] = msg
			l.cacheIndex++
			if l.cacheIndex == l.max {
				// TODO : max size of cached msg,flush into io
				if err := l.flush(); err != nil {
					fmt.Println(err)
				}

				// reset l.cache.'cause point,range may leak
				l.cache[0] = nil
				l.cacheIndex = 0
			}
		case <-l.done:
			debug_p("in done")
			// TODO: flush into io
			l.flush()
			for i := len(l.msgChan); i > 0; i++ {
				m := <-l.msgChan
				fmt.Fprint(l.w, l.fHead.FormatHead(), m.c)
			}
			close(l.msgChan)
			l = nil
			return

		}
	}
}

// flush just write to io
func (l *Logger) flush() error {
	debug_p(l.cache)
	if l.cache[0] == nil {
		return nil
	}
	var errString string

	for i := 0; i < l.cacheIndex; i++ {
		// debug_p("in loop i: ", i, " v : ", l.cache[i])
		debug_p(l.fHead.FormatHead(), l.cache[i].c)
		_, err := fmt.Fprintf(l.w, "%s %s", l.fHead.FormatHead(), l.cache[i].c)
		if err != nil {
			errString += err.Error()
		}
	}
	if errString != "" {
		return errors.New(errString)
	}
	return nil
}

/*
	setters of Logger,should done before run otherwise will not set
	the value and return false
*/
func (l *Logger) SetDebugFlag(b bool) bool {
	if l.inited {
		return false
	}
	l.debugFlag = b
	return true
}
func (l *Logger) SetMax(max int) bool {
	if l.inited {
		return false
	}
	l.max = max
	l.cache = make([]*msg, max)
	return true
}

// Debug
func (l *Logger) Debug(v interface{}) {
	if !l.debugFlag {
		return
	}

	l.msgChan <- &msg{
		c: v,
		t: DebugMsg,
	}

}

// Info
func (l *Logger) Info(v interface{}) {
	l.msgChan <- &msg{
		c: v,
		t: InfoMsg,
	}

}

// Error
func (l *Logger) Error(v interface{}) {
	l.msgChan <- &msg{
		c: v,
		t: ErrorMsg,
	}

}

// Stop
func (l *Logger) Stop() {
	l.done <- struct{}{}
	// time.Sleep(1 * time.Microsecond)
}

// cache

// need a format header
type FormatHeader interface {
	FormatHead() string
}

// default formatheader
type defautFormatHeader struct{}

// default format header : time + caller(depth = 0)
func (defautFormatHeader) FormatHead() string {
	t := time.Now().String()
	_, file, line, ok := runtime.Caller(0)
	if !ok {
		file = "???"
		line = -1
	}
	return fmt.Sprintf("%s %s %d ", t, file, line)
}

func (l *Logger) SetFHead(fHead FormatHeader) bool {
	if l.inited {
		return false
	}
	l.fHead = fHead
	return true
}

// func formatHead(s string)string {
// 	// t := time.Now()
// 	// return time.Now()
// }
