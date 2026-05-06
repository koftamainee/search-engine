package health

import (
	"net/http"

	"github.com/koftamainee/search-engine-common/pkg/response"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, nil)
	}
}
