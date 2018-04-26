package admin

import (
	"encoding/json"
	"net/http"
	"runtime"

	"godep.lzd.co/go-config"
	"godep.lzd.co/goprof"
	"godep.lzd.co/mobapi_lib/admin/sessionlogger"
	"godep.lzd.co/mobapi_lib/logger"
	"godep.lzd.co/mobapi_lib/resources"
)

func init() {
	config.RegisterBool("pprof-enabled", "Profiling tools enabled", true)
}

func NewHTTPHandler(serviceID, version, venture, env string, resources []resources.IResource,
	swaggerJSONHandler http.Handler, clientGenHandler http.Handler) http.Handler {

	pprofEnabled, _ := config.GetBool("pprof-enabled")

	mux := http.NewServeMux()

	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service_id":   serviceID,
			"version":      version,
			"venture":      venture,
			"env":          env,
			"pprofEnabled": pprofEnabled,
			"rps":          RPSCounter.Rate(),
		})
	})

	mux.Handle("/status/",
		http.StripPrefix("/status", NewStatusHTTPHandler(serviceID, resources)))

	if logsPath, _ := config.GetString("sessions-logs-path"); logsPath != "" {
		mux.Handle("/logs/", http.StripPrefix("/logs",
			sessionlogger.NewViewerHTTPHandler(sessionlogger.NewViewer(logsPath))))
	}

	mux.Handle("/swagger.json", swaggerJSONHandler)
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

	return mux
}
