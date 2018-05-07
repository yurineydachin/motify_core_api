package log

import (
	"testing"
	"time"
	"motify_core_api/godep_libs/go-log/format"
	"log/syslog"
	"bytes"
)

func BenchmarkGoLog(b *testing.B) {
	result := make(chan int)
	go createSocket(-1, time.Second*10, 0, false, result)
	<-result

	l := NewLogger("test-api", "unixgram", "/tmp/socket", DEBUG)
	f := &format.Std{Message: "message"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Record(DEBUG, f)
	}
}

func BenchmarkStdSyslog(b *testing.B) {
	result := make(chan int)
	go createSocket(-1, time.Second*10, 0, false, result)
	<-result

	l, _ := syslog.Dial("unixgram", "/tmp/socket", syslog.LOG_LOCAL2, "")
	f := &format.Std{Message: "message"}
	f.SetLevel(DEBUG)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		f.WriteTo(buf)
		l.Write(buf.Bytes())
	}
}
