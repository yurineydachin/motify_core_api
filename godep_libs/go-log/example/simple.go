package main

import (
	"flag"
	"net"
	"os"
	"sync/atomic"
	"time"

	"motify_core_api/godep_libs/go-log"
)

var network = flag.String("network", "unixgram", `Known networks are:
		"tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only),
		"udp", "udp4" (IPv4-only), "udp6" (IPv6-only),
		"ip", "ip4" (IPv4-only), "ip6" (IPv6-only),
		"unix", "unixgram" and "unixpacket"`)
var service = flag.String("service", "service_test", "Service name")
var address = flag.String("address", "/dev/log", "Address of syslog daemon")
var pause = flag.Duration("pause", time.Second, "Pause between messages")
var rt = flag.Duration("rt", 60*time.Second, "Run time")
var c = flag.Uint64("c", 1, "Concurrency level")
var socket = flag.String("socket", "", "Create socket file")
var bufferSize = flag.Int64("buffer", 32*1000*1024, "Buffer size in bytes")
var workers = flag.Int("workers", 4, "Worker count")

func main() {
	flag.Parse()
	done := make(chan bool)
	if *socket != "" {
		go createSocket(*socket, done)
	}

	l := log.New(log.WithServiceName(*service), log.WithDialOptions(*network, *address),
		log.WithLevel(log.INFO), log.WithBufferSize(*bufferSize), log.WithWorkerCount(*workers))

	var i uint64 = 0
	var stop uint64 = 0
	for j := 0; j < int(*c); j++ {
		go func(j int) {
			//fmt.Println("Started", j)
			for {
				s := atomic.LoadUint64(&stop)
				if s > 0 {
					return
				}

				l.Logf(0, log.Span{}, log.INFO, "tick %d", nil, atomic.AddUint64(&i, 1))
				time.Sleep(*pause)
			}
		}(j)
	}

	time.Sleep(*rt)
	atomic.StoreUint64(&stop, 1)
	l.Logf(0, log.Span{}, log.INFO, "done %d", nil, atomic.LoadUint64(&i))
	l.Close()
	//fmt.Println("Scheduled", atomic.LoadUint64(&i))
	//sent, lost, totalLoggedCount, totalLoggedLength, createdBuffersCount, skippedBuf := l.Stats()
	//fmt.Println("Sent:", sent, "Lost:", lost, "BuffersCount:", totalLoggedCount, "TotalBufferLength:", totalLoggedLength, "CreatedBuffers:", createdBuffersCount, "SkippedBuffers:", skippedBuf)
	if *socket != "" {
		done <- true
		<-done
	}
}

func createSocket(a string, stop chan bool) {
	addr, err := net.ResolveUnixAddr("unixgram", a)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		panic(err)
	}

	defer func() {
		conn.Close()
		os.Remove(a)
		close(stop)
	}()

	// read from socket
	go func() {
		for {
			buf := make([]byte, 256)
			conn.ReadFromUnix(buf)
			//println(string(buf))
		}
	}()

	<-stop
}
