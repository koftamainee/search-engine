package recoverer

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
)

func Middleware() middleware.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic recovered: %v\n%s", err, debug.Stack())
					response.Internal(w)
				}
			}()
			next(w, r)
		}
	}
}
