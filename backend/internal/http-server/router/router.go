package router

import (
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/login"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/logout"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/register"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/me"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/search"
	"github.com/koftamainee/search-engine/backend/internal/http-server/middleware"
	authmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	loggermw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/logger"
	recoverermv "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/recoverer"
	requestidmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/requestid"
	"github.com/koftamainee/search-engine/backend/internal/service"
)

func New(authService *service.AuthService, searchService *service.SearchService) http.Handler {

	mux := http.NewServeMux()

	recoverer := recoverermv.Middleware()
	requestid := requestidmw.Middleware()
	logger := loggermw.Middleware()
	auth := authmw.Middleware(authService)

	registerFunc := middleware.Chain(register.New(authService), recoverer, requestid, logger)
	loginFunc := middleware.Chain(login.New(authService), recoverer, requestid, logger)
	logoutFunc := middleware.Chain(logout.New(authService), recoverer, requestid, logger, auth)
	meFunc := middleware.Chain(me.New(), recoverer, requestid, logger, auth)

	searchFunc := middleware.Chain(search.New(searchService), recoverer, requestid, logger, auth)

	mux.HandleFunc("POST /v1/auth/register", registerFunc)
	mux.HandleFunc("POST /v1/auth/login", loginFunc)
	mux.HandleFunc("POST /v1/auth/logout", logoutFunc)

	mux.HandleFunc("GET /v1/me", meFunc)

	mux.HandleFunc("GET /v1/search", searchFunc)

	return mux
}
