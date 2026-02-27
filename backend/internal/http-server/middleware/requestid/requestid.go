package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware"
)

type contextKey string

const IDKey contextKey = "request_id"

func Middleware() middleware.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			id := generate()
			w.Header().Set("X-Request-ID", id)
			ctx := context.WithValue(r.Context(), IDKey, id)
			next(w, r.WithContext(ctx))
		}
	}
}

func GetID(r *http.Request) string {
	id, _ := r.Context().Value(IDKey).(string)
	return id
}

func generate() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
