package health

import (
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, nil)
	}
}
