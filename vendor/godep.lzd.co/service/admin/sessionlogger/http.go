package sessionlogger

import (
	"encoding/json"
	"net/http"
)

type ViewerHTTPHandler struct {
	viewer *Viewer
	mux    *http.ServeMux
}

func NewViewerHTTPHandler(viewer *Viewer) *ViewerHTTPHandler {
	handler := &ViewerHTTPHandler{
		viewer: viewer,
		mux:    http.NewServeMux(),
	}

	handler.mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		files, err := handler.viewer.GetLogsNames()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})

	handler.mux.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid trace id"))
			return
		}

		session, err := handler.viewer.GetSession(r.FormValue("file"), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)
	})

	return handler
}

func (h *ViewerHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
