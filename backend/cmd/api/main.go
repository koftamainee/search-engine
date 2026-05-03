package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/koftamainee/search-engine-common/pkg/logger"
	"github.com/koftamainee/search-engine/backend/internal/config"
	"github.com/koftamainee/search-engine/backend/internal/http-server/router"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
	"github.com/koftamainee/search-engine/backend/internal/service/search"
	"github.com/koftamainee/search-engine/backend/internal/service/suggest"
	"github.com/koftamainee/search-engine/backend/internal/storage/postgres"
	"github.com/koftamainee/search-engine/backend/internal/storage/redis"
	"github.com/meilisearch/meilisearch-go"
	redissdk "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := config.MustLoad()

	logCfg := logger.Config{
		Level:     slog.LevelInfo,
		JSON:      cfg.Env == "prod",
		AddSource: false,
		Color:     cfg.Env != "prod",
	}

	log := logger.New(logCfg)
	logger.SetDefault(log)

	log.Info("Starting backend service")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.Timeout)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	const connectionAttempts = 5

	var pgPool *pgxpool.Pool
	wg.Go(func() error {
		return connect(ctx, connectionAttempts, cfg.HTTPServer.Timeout, func() error {
			var err error
			pgPool, err = postgres.New(ctx, cfg.Postgres.URL)
			return err
		})
	})

	var meiliClient meilisearch.ServiceManager
	wg.Go(func() error {
		meiliClient = meilisearch.New(cfg.Meilisearch.URL, meilisearch.WithAPIKey(cfg.Meilisearch.ApiKey))

		return connect(ctx, connectionAttempts, cfg.HTTPServer.Timeout, func() error {
			_, err := meiliClient.Health()
			return err
		})
	})

	var redisClient *redissdk.Client
	wg.Go(func() error {
		return connect(ctx, connectionAttempts, cfg.HTTPServer.Timeout, func() error {
			var err error
			redisClient, err = redis.New(ctx, cfg.SessionStorage.Address, cfg.SessionStorage.Password, cfg.SessionStorage.DB)
			return err
		})
	})

	err := wg.Wait()
	if err != nil {
		log.Error("initialization failed", "error", err)
		os.Exit(1)
	}

	defer pgPool.Close()
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Warn("failed to close redis connection", "error", err)
		}
	}()

	meiliIndex := meiliClient.Index(cfg.Meilisearch.Index)

	userStorage := postgres.NewUserStorage(pgPool)
	historyStorage := postgres.NewHistoryStorage(pgPool)
	sessionStorage := redis.NewSessionStorage(redisClient)

	authService := auth.New(userStorage, sessionStorage, cfg.Env == "prod")
	searchService := search.New(meiliIndex, historyStorage)
	suggestService := suggest.New(meiliIndex, historyStorage)

	r := router.New(log, authService, searchService, suggestService)

	log.Info("starting server", "addr", cfg.HTTPServer.Address)

	err = http.ListenAndServe(cfg.HTTPServer.Address, r)
	if err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}

func connect(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	var err error

	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}
