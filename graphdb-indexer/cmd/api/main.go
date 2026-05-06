package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koftamainee/search-engine-common/pkg/logger"
	"github.com/koftamainee/search-engine/graphdb-indexer/internal/config"
	"github.com/koftamainee/search-engine/graphdb-indexer/internal/http-server/router"
	"github.com/nats-io/nats.go"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const (
	connectionTimeout  = 10 * time.Second // TODO load timeouts from ENV ??
	connectionAttempts = 5

	shutdownTimeout = 30 * time.Second
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(logger.Config{
		Level:     slog.LevelInfo,
		JSON:      cfg.Env == "prod",
		AddSource: false,
		Color:     cfg.Env != "prod",
	})
	logger.SetDefault(log)

	natsClient, neo4jClient, redisClient := mustConnectDB(cfg, log)

	mux := router.New(log)

	// TODO: maybe load from ENV
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("Starting graphdb-indexer", "address", cfg.HTTPServer.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit

	shutdown(natsClient, neo4jClient, redisClient, srv, log)
	fmt.Println("see you soon~")
}

func shutdown(natsClient *nats.Conn, neo4jClient neo4j.Driver,
	redisClient *redis.Client, srv *http.Server, log *slog.Logger) {
	log.Info("shutting down gracefully...", "timeout", shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown error", "error", err)
	} else {
		log.Info("http server stopped gracefully")
	}

	log.Info("closing connections...")

	if natsClient != nil {
		natsClient.Close()
		log.Info("NATS connection closed")
	}

	if neo4jClient != nil {
		if err := neo4jClient.Close(ctx); err != nil {
			log.Error("neo4j close error", "error", err)
		} else {
			log.Info("neo4j connection closed")
		}
	}

	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Error("redis close error", "error", err)
		} else {
			log.Info("redis connection closed")
		}
	}
}

func mustConnectDB(cfg *config.Config, log *slog.Logger) (*nats.Conn, neo4j.Driver, *redis.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	var natsClient *nats.Conn
	wg.Go(func() error {
		log.Info("connecting to message broker", "url", cfg.MessageBroker.URL)

		err := connect(ctx, connectionAttempts, 2*time.Second, func() error {
			conn, err := nats.Connect(cfg.MessageBroker.URL)
			if err != nil {
				log.Warn("message broker connection attempt failed", "error", err)
				return err
			}

			natsClient = conn
			return nil
		})

		if err != nil {
			log.Error("failed to connect to message broker", "error", err)
			return err
		}

		log.Info("connected to message broker")
		return nil
	})

	var neo4jClient neo4j.Driver
	wg.Go(func() error {
		log.Info("connecting to graphdb", "uri", cfg.GraphDB.URI)

		err := connect(ctx, connectionAttempts, 2*time.Second, func() error {
			driver, err := neo4j.NewDriver(
				cfg.GraphDB.URI,
				neo4j.BasicAuth(cfg.GraphDB.User, cfg.GraphDB.Pass, ""),
			)
			if err != nil {
				log.Warn("graphdb driver init failed", "error", err)
				return err
			}

			if err := driver.VerifyConnectivity(ctx); err != nil {
				log.Warn("graphdb connection failed", "error", err)
				return err
			}

			neo4jClient = driver
			return nil
		})

		if err != nil {
			log.Error("failed to connect to graphdb", "error", err)
			return err
		}

		log.Info("connected to graphdb")
		return nil
	})

	var redisClient *redis.Client
	wg.Go(func() error {
		log.Info("connecting to feature storage", "addr", cfg.FeatureStorage.Addr)

		err := connect(ctx, connectionAttempts, 2*time.Second, func() error {
			client := redis.NewClient(&redis.Options{
				Addr:     cfg.FeatureStorage.Addr,
				Password: cfg.FeatureStorage.Pass,
				DB:       0,
			})

			if err := client.Ping(ctx).Err(); err != nil {
				log.Warn("feature storage ping failed", "error", err)
				return err
			}

			redisClient = client
			return nil
		})

		if err != nil {
			log.Error("failed to connect to feature storage", "error", err)
			return err
		}

		log.Info("connected to feature storage")
		return nil
	})

	if err := wg.Wait(); err != nil {
		log.Error("initialization failed", "error", err)
		os.Exit(1)
	}
	return natsClient, neo4jClient, redisClient
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
