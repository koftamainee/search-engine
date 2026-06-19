package router

import (
	"log/slog"
	"net/http"

	"github.com/koftamainee/search-engine-common/pkg/middleware"
	logctxmw "github.com/koftamainee/search-engine-common/pkg/middleware/logctx"
	loggermw "github.com/koftamainee/search-engine-common/pkg/middleware/logger"
	recoverermw "github.com/koftamainee/search-engine-common/pkg/middleware/recoverer"
	requestidmw "github.com/koftamainee/search-engine-common/pkg/middleware/requestid"
	"github.com/koftamainee/search-engine/graphdb-indexer/internal/http-server/handlers/health"
)

func New(log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	recoverer := recoverermw.Middleware()
	logctx := logctxmw.Middleware(log)
	requestid := requestidmw.Middleware()
	logger := loggermw.Middleware()

	healthHandler := middleware.Chain(health.New(), recoverer, logctx, requestid, logger)

	mux.HandleFunc("GET /health", healthHandler)

	// TODO: add POST /index endpoint
	// TODO: It should launch a async task which will calculate metrics on graphdb and post it to feature storage
	// TODO: tracking system will be greatly appreciated. Webhook or something

	// NOTE: /index and other POST endpoints require auth.
	// NOTE: @koftamainee should implement JWT-based auth system and middleware in search-engine-common

	return mux
}
