package main

import (
	"context"
	"log"
	"net/http"

	config "github.com/koftamainee/search-engine/backend/internal/config"
	"github.com/koftamainee/search-engine/backend/internal/http-server/router"
	"github.com/koftamainee/search-engine/backend/internal/service"
	"github.com/koftamainee/search-engine/backend/internal/storage/postgres"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()

	pool, err := postgres.New(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	userStorage := postgres.NewUserStorage(pool)
	sessionStorage := postgres.NewSessionStorage(pool)

	authService := service.NewAuthService(userStorage, sessionStorage)

	r := router.New(authService)

	log.Printf("starting server on %s", cfg.HTTPServer.Address)
	err = http.ListenAndServe(cfg.HTTPServer.Address, r)
	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
