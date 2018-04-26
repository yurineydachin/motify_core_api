package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/sergei-svistunov/gorpc/transport/http_json"
	"motify_core_api/godep_libs/goprof"
	logger_viewer "motify_core_api/godep_libs/service/admin/sessionlogger"
	"motify_core_api/godep_libs/service/admin/static"
	"motify_core_api/godep_libs/service/config"
	"motify_core_api/godep_libs/service/dconfig"
	"motify_core_api/godep_libs/service/interfaces"
	"motify_core_api/godep_libs/service/k8s"
	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/godep_libs/swgui"
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
	config.RegisterBool("pprof-enabled", "Profiling tools enabled", false)
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
				if host, err := k8s.GetHostname(); err == nil {
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

func NewHTTPHandler(serviceID, version, venture, env string, resources []interfaces.IResource, dconf *dconfig.Manager,
	swaggerJSONCallbacks http_json.SwaggerJSONCallbacks, swaggerUIHandler, swaggerJSONHandler, clientGenHandler http.Handler) http.Handler {

	pprofEnabled, _ := config.GetBool("pprof-enabled")

	mux := http.NewServeMux()

	mux.Handle("/", static.NewHTTPHandler())
	//mux.Handle("/", http.FileServer(http.Dir("./admin/static/html")))

	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service_id":   serviceID,
			"version":      version,
			"venture":      venture,
			"env":          env,
			"pprofEnabled": pprofEnabled,
		})
	})

	mux.Handle("/status/",
		http.StripPrefix("/status",
			accessHandler(NewStatusHTTPHandler(serviceID, resources))))

	mux.Handle("/settings/",
		http.StripPrefix("/settings",
			accessHandler(NewDSettingsHTTPHandler(dconf))))
	mux.HandleFunc("/login/", login)

	logsPath, _ := config.GetString("sessions_logs_path")
	if logsPath != "" {
		mux.Handle("/logs/", http.StripPrefix("/logs", logger_viewer.NewViewerHTTPHandler(logger_viewer.NewViewer(logsPath))))
	}

	mux.Handle("/swagger.json", swaggerJSONHandler)

	if _, ok := swaggerUIHandler.(*swgui.Handler); ok {
		mux.Handle("/docs/", swaggerUIHandler)
		//because swaggerUIHandler use swagger json generator that work on "/docs/swagger.json" route
		mux.Handle("/docs/swagger.json", swaggerJSONHandler)
	} else {
		mux.Handle("/docs/", http.StripPrefix("/docs", swaggerUIHandler))
	}
	mux.Handle("/client_sdk_gen/", http.StripPrefix("/client_sdk_gen", clientGenHandler))

	if pprofEnabled {
		goprof.SetLogFunction(func(format string, args ...interface{}) {
			logger.Warning(nil, format, args...)
		})
		mux.Handle("/pprof/", http.StripPrefix("/pprof", goprof.NewHandler()))
		autoProfiling(mux)
	} else {
		runtime.MemProfileRate = 0
	}

	mux.HandleFunc("/config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(config.JSON())
	})

	return authHandler(mux)
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
