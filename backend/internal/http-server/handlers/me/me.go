package me

import (
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(auth.UserContextKey).(*domain.User)
		if !ok {
			response.Unauthorized(w)
			return
		}

		response.OK(w, user)
	}
}
