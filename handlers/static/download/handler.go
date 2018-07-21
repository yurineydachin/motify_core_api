package download

import (
	"net/http"
)

type Handler struct {
	urlPath, dirPath string
}

func New(urlPath, dirPath string) *Handler {
	return &Handler{
		urlPath: urlPath,
		dirPath: dirPath,
	}
}

func (h *Handler) GetHttpHandler() http.Handler {
	return http.StripPrefix(h.urlPath, http.FileServer(http.Dir(h.dirPath)))
}
