package proxy

import (
	"io"
	"net/http"
	"time"
)

type Handler struct {
	target string
	client *http.Client
}

func New(target string, timeout time.Duration) *Handler {
	return &Handler{
		target: target,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, _ := http.NewRequest("GET", h.target+r.URL.RequestURI(), nil)
	resp, err := h.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for header, values := range src {
		for _, value := range values {
			dst.Set(header, value)
		}
	}
}
