package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/koftamainee/search-engine/backend/internal/config"
	"github.com/koftamainee/search-engine/backend/internal/http-server/router"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
	"github.com/koftamainee/search-engine/backend/internal/service/search"
	"github.com/koftamainee/search-engine/backend/internal/storage/postgres"
	"github.com/koftamainee/search-engine/backend/internal/storage/redis"
	"github.com/meilisearch/meilisearch-go"
	redis2 "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

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

func main() {
	cfg := config.MustLoad()

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
	defer pgPool.Close()

	var meiliClient meilisearch.ServiceManager
	wg.Go(func() error {
		meiliClient = meilisearch.New(cfg.Meilisearch.URL, meilisearch.WithAPIKey(cfg.Meilisearch.ApiKey))

		return connect(ctx, connectionAttempts, cfg.HTTPServer.Timeout, func() error {
			_, err := meiliClient.Health()
			return err
		})
	})

	var redisClient *redis2.Client

	wg.Go(func() error {
		return connect(ctx, connectionAttempts, cfg.HTTPServer.Timeout, func() error {
			var err error
			redisClient, err = redis.New(ctx, cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
			return err
		})
	})

	err := wg.Wait()
	if err != nil {
		log.Fatalf("initialization failed: %v", err)
	}

	defer pgPool.Close()
	defer func(redisClient *redis2.Client) {
		err := redisClient.Close()
		if err != nil {
			log.Printf("failed to close redis connection")
		}
	}(redisClient)

	meiliIndex := meiliClient.Index(cfg.Meilisearch.Index)

	userStorage := postgres.NewUserStorage(pgPool)
	sessionStorage := redis.NewSessionStorage(redisClient)

	authService := auth.New(userStorage, sessionStorage, cfg.Env == "prod")
	searchService := search.New(meiliIndex)

	r := router.New(authService, searchService)

	log.Printf("starting server on %s", cfg.HTTPServer.Address)
	err = http.ListenAndServe(cfg.HTTPServer.Address, r)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
