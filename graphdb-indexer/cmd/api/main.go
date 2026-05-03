package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/koftamainee/search-engine-common/pkg/logger"
	"github.com/koftamainee/search-engine-common/pkg/middleware"
	"github.com/koftamainee/search-engine-common/pkg/middleware/logctx"
	loggermw "github.com/koftamainee/search-engine-common/pkg/middleware/logger"
	"github.com/koftamainee/search-engine-common/pkg/middleware/recoverer"
	"github.com/koftamainee/search-engine-common/pkg/middleware/requestid"
	"github.com/koftamainee/search-engine-common/pkg/response"
)

func main() {

	log := logger.New(logger.Config{
		Level:     slog.LevelInfo,
		JSON:      false,
		AddSource: false,
		Color:     true,
	})
	logger.SetDefault(log)

	port := os.Getenv("GRAPHDB_INDEXER_PORT")
	if port == "" {
		log.Warn("No port defined in GRAPHDB_INDEXER_PORT, using 8080")
		port = "8080"
	}

	log.Info("Starting graphdb-indexer", "port", port)

	mux := http.NewServeMux()

	healthFunc := func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, nil)
	}

	healthHandler := middleware.Chain(healthFunc, recoverer.Middleware(), logctx.Middleware(log), requestid.Middleware(), loggermw.Middleware())

	mux.HandleFunc("GET /health", healthHandler)

	// TODO: maybe load from ENV
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
