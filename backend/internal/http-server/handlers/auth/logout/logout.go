package logout

import (
	"errors"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	libtoken "github.com/koftamainee/search-engine/backend/internal/lib/api/token"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
)

func New(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := libtoken.Extract(r)

		if token == "" {
			response.Unauthorized(w)
			return
		}

		err := authService.Logout(r.Context(), token)
		if err != nil {
			if errors.Is(err, auth.ErrNotFound) {
				response.Unauthorized(w)
				return
			}
			response.Internal(w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Secure:   authService.IsProd,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		})

		response.OK(w, nil)
	}
}
