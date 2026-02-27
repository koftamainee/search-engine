package logger

import (
	"log"
	"net/http"
	"time"

	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware/requestid"
)

func Middleware() middleware.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next(wrapped, r)

			log.Printf(
				"| %d | %s | %s | %s | %s | %s | %s |",
				wrapped.status,
				time.Since(start),
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				r.UserAgent(),
				requestid.GetID(r),
			)
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if !rw.wroteHeader {
		rw.status = status
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(status)
	}
}
