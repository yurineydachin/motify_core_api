package google

import (
	"context"
)

// v1Args contains a request arguments
type v1Args struct {
	Query string `key:"query" description:"Search query"`
}

// V1Res is a response
type V1Res struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Errors defining and binding
type v1ErrorTypes struct {
	EMPTY_RESULT error `text:"Nothing was found"`
}

var v1Errors v1ErrorTypes

func (*Handler) V1ErrorsVar() *v1ErrorTypes {
	return &v1Errors
}

// V1 is a version 1 implementation of this handler
func (h *Handler) V1(ctx context.Context, opts *v1Args) ([]V1Res, error) {
	seRes, err := h.se.Search(ctx, opts.Query)
	if err != nil {
		return nil, err
	}

	if len(seRes) == 0 {
		return nil, v1Errors.EMPTY_RESULT
	}

	result := make([]V1Res, len(seRes))
	for i, row := range seRes {
		result[i].Title = row.Title
		result[i].URL = row.URL
	}

	return result, nil
}
