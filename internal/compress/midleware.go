package compress

import (
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/logger"
)

// ContentTypesToBeEncoded defince the content types to be encoded
var ContentTypesToBeDecoded = []string{"application/json", "text/html"}

// Middleware for decompression
func Middleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			logger.Log.Infow(
				"Decoding content from gzip",
				"uri", r.RequestURI,
				"method", r.Method,
			)
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
