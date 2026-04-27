package register

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	auth "github.com/koftamainee/search-engine/backend/internal/service/auth"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func New(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.BadRequest(w, "invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			response.BadRequest(w, "email and password are required")
			return
		}

		user, err := authService.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrAlreadyExists) {
				response.BadRequest(w, "user with this email already exists")
				return
			}
			response.Internal(w)
			return
		}

		response.Created(w, Response{
			ID:    user.ID,
			Email: user.Email,
		})
	}
}
