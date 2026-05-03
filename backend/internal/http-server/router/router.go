package router

import (
	"log/slog"
	"net/http"

	"github.com/koftamainee/search-engine-common/pkg/middleware"
	logctxmw "github.com/koftamainee/search-engine-common/pkg/middleware/logctx"
	loggermw "github.com/koftamainee/search-engine-common/pkg/middleware/logger"
	recoverermv "github.com/koftamainee/search-engine-common/pkg/middleware/recoverer"
	requestidmw "github.com/koftamainee/search-engine-common/pkg/middleware/requestid"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/health"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/login"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/logout"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/auth/register"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/me"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/search"
	"github.com/koftamainee/search-engine/backend/internal/http-server/handlers/suggest"
	authmw "github.com/koftamainee/search-engine/backend/internal/http-server/middleware/auth"
	authService "github.com/koftamainee/search-engine/backend/internal/service/auth"
	searchService "github.com/koftamainee/search-engine/backend/internal/service/search"
	suggestService "github.com/koftamainee/search-engine/backend/internal/service/suggest"
)

func New(log *slog.Logger, authService *authService.Service, searchService *searchService.Service, suggestService *suggestService.Service) http.Handler {

	mux := http.NewServeMux()

	recoverer := recoverermv.Middleware()
	requestid := requestidmw.Middleware()
	logctx := logctxmw.Middleware(log)
	logger := loggermw.Middleware()
	auth := authmw.Middleware(authService)

	registerFunc := middleware.Chain(register.New(authService), recoverer, logctx, requestid, logger)
	loginFunc := middleware.Chain(login.New(authService), recoverer, logctx, requestid, logger)
	logoutFunc := middleware.Chain(logout.New(authService), recoverer, logctx, requestid, logger, auth)

	healthFunc := middleware.Chain(health.New(), recoverer, logctx, requestid, logger)

	meFunc := middleware.Chain(me.New(), recoverer, logctx, requestid, logger, auth)

	searchFunc := middleware.Chain(search.New(searchService), recoverer, logctx, requestid, logger, auth)
	suggestFunc := middleware.Chain(suggest.New(suggestService), recoverer, logctx, requestid, logger, auth)

	mux.HandleFunc("POST /v1/auth/register", registerFunc)
	mux.HandleFunc("POST /v1/auth/login", loginFunc)
	mux.HandleFunc("POST /v1/auth/logout", logoutFunc)

	mux.HandleFunc("GET /v1/health", healthFunc)

	mux.HandleFunc("GET /v1/me", meFunc)

	mux.HandleFunc("GET /v1/search", searchFunc)
	mux.HandleFunc("GET /v1/suggest", suggestFunc)

	return mux
}
