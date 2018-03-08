package admin

import (
	"net/http"

	"godep.lzd.co/service/logger"
)

func CheckCertificate(certificates []string, h http.Handler) http.Handler {
	if len(certificates) == 0 {
		logger.Notice(nil, "No certificates allowed.")
	}

	certificateMatches := make([]string, 0, len(certificates))
	for _, certificate := range certificates {
		certificateMatches = append(certificateMatches, `/CN=`+certificate+`/`)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
