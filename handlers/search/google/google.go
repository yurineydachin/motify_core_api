package google

import "motify_core_api/resources/searchengine"

type Handler struct {
	se *searchengine.SearchEngine
}

func New(se *searchengine.SearchEngine) *Handler {
	return &Handler{
		se: se,
	}
}

func (*Handler) Caption() string {
	return "Google search"
}

func (*Handler) Description() string {
	return "Search things in Lazada using Google"
}
