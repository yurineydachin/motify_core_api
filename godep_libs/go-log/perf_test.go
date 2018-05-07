package log

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"motify_core_api/godep_libs/go-log/format"
	"math"
)

func createSocket(targetMsgCount int, timeout time.Duration, slowReadFactor int, debug bool, result chan int) {
	os.Remove("/tmp/socket")

	addr, err := net.ResolveUnixAddr("unixgram", "/tmp/socket")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		panic(err)
	}

	defer func() {
		conn.Close()
		os.Remove("/tmp/socket")
	}()

	result <- 0

	//read from socket
	var readCount int
	if targetMsgCount == -1 {
		targetMsgCount = math.MaxInt64
	}
	if targetMsgCount >= 0 {
		buf := make([]byte, 1024)
		timeout := time.After(timeout)

	readloop:
		for readCount = 0; readCount < targetMsgCount; {
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
			_, _, err := conn.ReadFromUnix(buf)
			if err == nil && buf[0] == 60 { // 60 = `<`
				buf[0] = 0
				readCount++
				if debug {
					fmt.Print(string(buf))
				}
				if slowReadFactor > 0 && readCount % slowReadFactor == 0 {
					time.Sleep(1)
				}
			}

			select {
			case <-timeout:
				break readloop
			default:

			}
		}
	}

	if result != nil {
		result <- readCount
	}

}

func mustSend(l *Logger, msg int, concurrency int, duration time.Duration) (success bool, errorsCount int64) {
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	var e int64
	for c := 0; c < concurrency; c++ {
		go func(c int) {
			n := msg / concurrency
			if concurrency-1 == c {
				n += msg % concurrency
			}
			for i := 0; i < n; i++ {
				if err := l.Record(DEBUG, &format.Std{Message: "message"}); err != nil {
					atomic.AddInt64(&e, 1)
				}
			}
			wg.Done()
		}(c)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true, atomic.LoadInt64(&e)
	case <-time.After(duration):
		return false, atomic.LoadInt64(&e)
	}
}

func testPerfWithoutSocketReader(t *testing.T) {
	result := make(chan int)
	go createSocket(0, time.Second*1, 0, false, result)
	<-result

	n, c, d := 30000, 100, 1*time.Second

	l := New(WithServiceName("test-perf-wo-socket"), WithDialOptions("unixgram", "/tmp/socket"), WithErrorWriter(false))
	ok, _ := mustSend(l, n, c, d)
	if !ok  {
		t.Errorf("expected %d messages with %d concurrency was sent to socket per %s", n, c, d)
	}
}

func testPerfWithoutSocketReaderWithErrorWriter(t *testing.T) {
	result := make(chan int)
	go createSocket(0, time.Second*1, 0, false, result)
	<-result

	n, c, d := 3000, 100, 1*time.Second

	l := New(WithServiceName("test-perf-wo-socket-with-err"), WithDialOptions("unixgram", "/tmp/socket"),
		WithErrorWriter(true), WithBufferSize(0))
	ok, e := mustSend(l, n, c, d)
	if !ok  {
		t.Errorf("expected %d messages with %d concurrency was sent to socket per %s", n, c, d)
	}
	if e != int64(n) {
		t.Errorf("expected errors %d, got: %d", n, e)
	}
}

func testPerfWithSocketReader(t *testing.T) {
	messages := 30000
	loggers := 10
	lost := 30
	result := make(chan int)

	go createSocket(messages, time.Second*1, 0, false, result)
	<-result

	wg := sync.WaitGroup{}
	wg.Add(loggers)
	for i := 0; i < loggers; i++ {
		go func(i int) {
			n := messages / loggers
			if loggers-1 == i {
				n += messages % loggers
			}
			c, d := 30, 1*time.Second

			l := New(WithServiceName("test-perf-with-err-"+strconv.Itoa(i)), WithDialOptions("unixgram", "/tmp/socket"), WithErrorWriter(true))
			ok, _ := mustSend(l, n, c, d)
			if !ok {
				t.Errorf("expected %d messages with %d concurrency was sent to socket per %s", n, c, d)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	if p := messages - <-result; p > lost {
		t.Errorf("expected lost messages <= %d (%f%%), got: %d (%f%%)", lost, float64(lost)/float64(messages)*100, p, float64(p)/float64(messages)*100)
	}
}

func testPerfWithSlowSocketReader(t *testing.T) {
	messages := 30000
	loggers := 10
	lost := 300
	result := make(chan int)

	go createSocket(messages, time.Second*1, 200, false, result)
	<-result

	wg := sync.WaitGroup{}
	wg.Add(loggers)
	for i := 0; i < loggers; i++ {
		go func(i int) {
			n := messages / loggers
			if loggers-1 == i {
				n += messages % loggers
			}
			c, d := 30, 1*time.Second

			l := New(WithServiceName("test-perf-wo-err-"+strconv.Itoa(i)), WithDialOptions("unixgram", "/tmp/socket"), WithErrorWriter(false))
			ok, _ := mustSend(l, n, c, d)
			if !ok {
				t.Errorf("expected %d messages with %d concurrency was sent to socket per %s", n, c, d)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	if p := messages - <-result; p > lost {
		t.Errorf("expected lost messages <= %d (%f%%), got: %d (%f%%)", lost, float64(lost)/float64(messages)*100, p, float64(p)/float64(messages)*100)
	}
}
