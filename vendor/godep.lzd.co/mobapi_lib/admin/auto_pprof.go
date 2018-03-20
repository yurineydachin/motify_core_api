package admin

import (
	"net/http"
	"time"
	"fmt"
	"godep.lzd.co/go-dconfig"
	"strings"
	"godep.lzd.co/mobapi_lib/utils/middleware"
)

const autoProfilingRPSCheckerInterval = 10 // seconds

var RPSCounter = NewCounter(autoProfilingRPSCheckerInterval * time.Second)
var autoProfilingType = "all"
var autoProfilingDuration = 30 * time.Second
var autoProfilingRPS uint = 0
var autoProfilingInterval = 24 * time.Hour
var autoProfilingLastRun time.Time
var autoProfilingEnabled = false

func init() {
	dconfig.RegisterString("auto-profiling-type", "Auto profiling type (values: all, cpu, heap, trace, goroutine, threadcreate, block)", autoProfilingType,
		func(v string) {
			switch v {
			case "all", "cpu", "heap", "trace", "goroutine", "threadcreate", "block":
				autoProfilingType = v
			default: //ignore other values
			}
		})
	dconfig.RegisterDuration("auto-profiling-duration", "Auto profiling duration (0 - disabled)", autoProfilingDuration,
		func(v time.Duration) {
			autoProfilingDuration = v
		})
	dconfig.RegisterUint("auto-profiling-rps", "RPS threshold which should be reached to enable auto profiling (0 - disabled)", autoProfilingRPS,
		func(v uint) {
			autoProfilingRPS = v
		})
	dconfig.RegisterDuration("auto-profiling-interval", "Minimum interval between two auto-profiling runs", autoProfilingInterval,
		func(v time.Duration) {
			autoProfilingInterval = v
			autoProfilingLastRun = time.Time{}
		})
	dconfig.RegisterString("auto-profiling-hosts", "Comma-separated list of hosts, where need to enable auto profiling (empty - disabled everywhere)", "",
		func(v string) {
			if v != "" {
				if host, err := middleware.GetHostname(); err == nil {
					list := strings.Split(v, ",")
					for _, h := range list {
						if h == host {
							autoProfilingEnabled = true
							return
						}
					}
				}
			}
			autoProfilingEnabled = false
		})
}

type fakeResponseWriter struct {
	header http.Header
}

func (r *fakeResponseWriter) Header() http.Header {
	return r.header
}

func (r *fakeResponseWriter) Write(d []byte) (int, error) {
	return 0, nil
}

func (r *fakeResponseWriter) WriteHeader(c int) {
	return
}

func autoProfiling(h http.Handler) {
	go func() {
		for {
			if autoProfilingEnabled && autoProfilingRPS > 0 && autoProfilingDuration > 0 &&
				time.Now().After(autoProfilingLastRun.Add(autoProfilingInterval)) &&
				RPSCounter.Rate()/autoProfilingRPSCheckerInterval >= uint64(autoProfilingRPS) {
				start, _ := http.NewRequest("GET", fmt.Sprintf("/pprof/toggle?enable=1&profile=%s", autoProfilingType), nil)
				h.ServeHTTP(&fakeResponseWriter{http.Header{}}, start)
				autoProfilingLastRun = time.Now()
				time.Sleep(autoProfilingDuration)
				stop, _ := http.NewRequest("GET", "/pprof/toggle?enable=0", nil)
				h.ServeHTTP(&fakeResponseWriter{http.Header{}}, stop)
			}
			time.Sleep(autoProfilingRPSCheckerInterval * time.Second)
		}
	}()
}

