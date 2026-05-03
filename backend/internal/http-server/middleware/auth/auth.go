package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/koftamainee/search-engine-common/pkg/middleware"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	libtoken "github.com/koftamainee/search-engine/backend/internal/lib/api/token"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func Middleware(authService *auth.Service) middleware.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := libtoken.Extract(r)
			if token == "" {
				response.Unauthorized(w)
				return
			}

			user, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, auth.ErrUserBanned) {
					response.Forbidden(w)
					return
				}
				response.Unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next(w, r.WithContext(ctx))
		}
	}
}
