package admin

import (
	"bytes"
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"

	"github.com/sergei-svistunov/gorpc/swagger_ui"
	"motify_core_api/godep_libs/service/logger"
)

type DocsHandler struct {
	swaggerUIHandler *swagger_ui.HTTPHandler
	postProcess      func([]byte) []byte
}

func NewDocsHandler(postProcess func([]byte) []byte) *DocsHandler {
	return &DocsHandler{
		swaggerUIHandler: swagger_ui.NewHTTPHandler(),
		postProcess:      postProcess,
	}
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// use default handler logic for non-home page requests
		h.swaggerUIHandler.ServeHTTP(w, r)
		return
	}

	mockWriter := httptest.NewRecorder()
	h.swaggerUIHandler.ServeHTTP(mockWriter, r)
	if mockWriter.Code != http.StatusOK {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if mockWriter.Body == nil || mockWriter.Body.Len() == 0 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// TODO: should we not check that content is gzipped?
	//	if mockWriter.Header().Get("Content-Encoding") == "gzip" {}
	html, err := gunzipIt(mockWriter.Body.Bytes())
	if err != nil {
		logger.Error(nil, "Can't gunzip swagger index html: %s", err)
		// use default handler logic
		h.swaggerUIHandler.ServeHTTP(w, r)
		return
	}

	// replace some content
	if h.postProcess != nil {
		html = h.postProcess(html)
	}

	compressed, err := gzipIt(html)
	if err != nil {
		logger.Error(nil, "Can't gzip swagger index html: %s", err)
		// send uncompressed page
		w.Header().Set("Content-Type", mime.TypeByExtension("html"))
		w.Write(html)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension("html"))
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(compressed)
}

func gunzipIt(data []byte) ([]byte, error) {
	in := bytes.NewReader(data)
	zipReader, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	var b bytes.Buffer
	_, err = io.Copy(&b, zipReader)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func gzipIt(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	w.Close()

	return b.Bytes(), nil
}
