package login

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/koftamainee/search-engine-common/pkg/middleware/logctx"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
)

type Request struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func New(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logctx.GetLogger(r)

		var req Request

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", "error", err)
			response.BadRequest(w, "invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			log.Warn("missing credentials", "email", req.Email)
			response.BadRequest(w, "email and password are required")
			return
		}

		session, err := authService.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				log.Warn("invalid credentials", "email", req.Email)
				response.Unauthorized(w)
				return
			}
			if errors.Is(err, auth.ErrUserBanned) {
				log.Warn("banned user attempted login", "email", req.Email)
				response.Forbidden(w)
				return
			}
			log.Error("login failed", "error", err)
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

		log.Info("user logged in", "email", req.Email)
		response.OK(w, nil)
	}
}
