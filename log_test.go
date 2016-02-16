package hlog

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

type testHead struct{}

func (testHead) FormatHead() string {
	return ""
}
func TestInfo(t *testing.T) {
	l := New(os.Stdout)
	l.Run()
	l.Info("hello world")

	l.Stop()
}

func TestDebug(t *testing.T) {
	const (
		msg = "debugInfo"
	)
	var tHead = testHead{}

	var testString = fmt.Sprintf("%s %s", tHead.FormatHead(), msg)

	var buf bytes.Buffer
	l := New(&buf)
	l.SetFHead(tHead)
	l.Run()
	l.Debug(msg)
	l.Stop()
	res := buf.String()
	if strings.TrimSpace(res) == strings.TrimSpace(testString) {
		t.Errorf("need %s but get %s \n", testString, res)
	}

	buf.Reset()
	ll := New(&buf)
	ll.SetFHead(tHead)
	ll.SetDebugFlag(true)
	ll.Run()
	ll.Debug(testString)
	if buf.String() != "" {
		t.Errorf("need %s but get %s \n", testString, buf.String())
	}
	ll.Stop()

}

func TestErr(t *testing.T) {
	var (
		msg = []string{"debugInfo"}
	)

	var tHead = testHead{}

	var testString = fmt.Sprintf("%v %v", tHead.FormatHead(), msg)
	// fmt.Println(testString)
	// fmt.Println("test head : ", tHead.FormatHead())
	var buf bytes.Buffer
	l := New(&buf)
	l.SetFHead(tHead)
	l.Run()
	l.Error(msg[0])
	l.Stop()
	result := buf.String()
	if strings.TrimSpace(result) != strings.TrimSpace(testString) {
		t.Errorf("need '%v' but get '%v' \n", testString, result)
	}
	// fmt.Println("result ::::::::: ", result)

}

func BenchmarkInfo(b *testing.B) {
	const testString = "test"
	b.StopTimer()
	var buf bytes.Buffer
	var mux sync.Mutex

	l := New(&buf)

	l.Run()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		go func() {
			mux.Lock()
			buf.Reset()
			mux.Unlock()
			for i := 0; i < 100; i++ {
				l.Info(testString)
			}
		}()
		go func() {
			mux.Lock()
			buf.Reset()
			mux.Unlock()
			for i := 0; i < 100; i++ {
				l.Info(testString)
			}
		}()

	}
	// l.Stop()

}

func TestSetMax(t *testing.T) {

	l := New(os.Stdout)
	l.Run()
	fmt.Println(l.max)
	l.Stop()

	ll := New(os.Stdout)
	ll.SetMax(1000)
	if ll.max != 1000 {
		t.Error("SetMax can done before run", ll.max)
	}
	ll.Run()
	ll.SetMax(50)
	if ll.max == 50 {
		t.Error("SetMax should done before Run", ll.max)
	}
	ll.Stop()
}
