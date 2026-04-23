package router

import (
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/login"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/logout"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/register"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware"
	authmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	corsmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/cors"
	loggermw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/logger"
	recoverermv "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/recoverer"
	requestidmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/requestid"
	"github.com/koftamainee/search-engine/backend/internal/service"
)

func New(authService *service.AuthService) http.Handler {

	mux := http.NewServeMux()

	recoverer := recoverermv.Middleware()
	requestid := requestidmw.Middleware()
	logger := loggermw.Middleware()
	cors := corsmw.Middleware()
	auth := authmw.Middleware(authService)

	registerFunc := middleware.Chain(register.New(authService), recoverer, requestid, logger, cors)
	loginFunc := middleware.Chain(login.New(authService), recoverer, requestid, logger, cors)
	logoutFunc := middleware.Chain(logout.New(authService), recoverer, requestid, logger, cors, auth)

	mux.HandleFunc("POST /v1/auth/register", registerFunc)
	mux.HandleFunc("POST /v1/auth/login", loginFunc)
	mux.HandleFunc("POST /v1/auth/logout", logoutFunc)

	return mux
}
