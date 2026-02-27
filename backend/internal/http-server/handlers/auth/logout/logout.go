package logout

import (
	"errors"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	libtoken "github.com/koftamainee/search-engine/backend/internal/lib/api/token"
	"github.com/koftamainee/search-engine/backend/internal/service"
)

func New(authService *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := libtoken.Extract(r)

		if token == "" {
			response.Unauthorized(w)
			return
		}

		err := authService.Logout(r.Context(), token)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				response.Unauthorized(w)
				return
			}
			response.Internal(w)
			return
		}

		response.OK(w, nil)
	}
}
