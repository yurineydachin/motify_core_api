package google

import "godep.lzd.co/mobapi_lib/_example/resources/searchengine"

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
