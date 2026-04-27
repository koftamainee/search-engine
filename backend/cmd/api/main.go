package main

import (
	"context"
	"log"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/config"
	"github.com/koftamainee/search-engine/backend/internal/http-server/router"
	"github.com/koftamainee/search-engine/backend/internal/service/auth"
	"github.com/koftamainee/search-engine/backend/internal/service/search"
	"github.com/koftamainee/search-engine/backend/internal/storage/postgres"
	"github.com/koftamainee/search-engine/backend/internal/storage/redis"
	"github.com/meilisearch/meilisearch-go"
	redis2 "github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer pool.Close()

	meiliClient := meilisearch.New(cfg.Meilisearch.URL, meilisearch.WithAPIKey(cfg.Meilisearch.ApiKey))
	meiliIndex := meiliClient.Index(cfg.Meilisearch.Index)

	redisClient, err := redis.New(ctx, cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis")
	}
	defer func(redisClient *redis2.Client) {
		err := redisClient.Close()
		if err != nil {
			log.Printf("Failed to close redis connection")
		}
	}(redisClient)

	userStorage := postgres.NewUserStorage(pool)
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
