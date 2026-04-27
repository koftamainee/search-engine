package login

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

		session, err := authService.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				response.Unauthorized(w)
				return
			}
			if errors.Is(err, auth.ErrUserBanned) {
				response.Forbidden(w)
				return
			}
			log.Printf("%v", err)

			response.Internal(w)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    session.Token,
			Secure:   authService.IsProd,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   int(auth.SessionDuration.Seconds()),
		})

		response.OK(w, nil)
	}
}
