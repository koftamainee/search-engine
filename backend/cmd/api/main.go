package main

import (
	"context"
	"log"
	"net/http"

	"github.com/koftamainee/search-engine/backend/internal/config"
	"github.com/koftamainee/search-engine/backend/internal/http-server/router"
	"github.com/koftamainee/search-engine/backend/internal/service"
	"github.com/koftamainee/search-engine/backend/internal/storage/postgres"
	"github.com/koftamainee/search-engine/backend/internal/storage/redis"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer pool.Close()

	redisClient, err := redis.New(ctx, cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis")
	}
	defer redisClient.Close()

	userStorage := postgres.NewUserStorage(pool)
	sessionStorage := redis.NewSessionStorage(redisClient)

	authService := service.NewAuthService(userStorage, sessionStorage)

	r := router.New(authService)

	log.Printf("starting server on %s", cfg.HTTPServer.Address)
	err = http.ListenAndServe(cfg.HTTPServer.Address, r)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
