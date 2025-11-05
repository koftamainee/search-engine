package router

import (
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/register"
	"github.com/koftamainee/search-engine/backend/internal/lib/api/response"
	"github.com/koftamainee/search-engine/backend/internal/service"
)

func New(authService *service.AuthService) http.Handler {

	mux := http.NewServeMux()

	//TODO: register all funcs

	mux.HandleFunc("GET /api/v1/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "Hello world!")
	})

	mux.HandleFunc("POST /api/v1/auth/register", register.New(authService))

	return mux
}
