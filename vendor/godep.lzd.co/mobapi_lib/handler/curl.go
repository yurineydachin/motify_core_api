package handler

import (
	"github.com/moul/http2curl"
	"net/http"
)

type Curl string

func (c Curl) String() string {
	return string(c)
}

func CurlFromRequest(r *http.Request) Curl {
	s, _ := http2curl.GetCurlCommand(r)
	return Curl(s.String())
}