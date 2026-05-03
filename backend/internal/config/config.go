package config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type HTTPServer struct {
	Address     string        `json:"address"`
	Timeout     time.Duration `json:"timeout"`
	IdleTimeout time.Duration `json:"idle_timeout"`
}

type Postgres struct {
	URL string `json:"url"`
}

type Meilisearch struct {
	URL    string `json:"url"`
	ApiKey string `json:"api_key"`
	Index  string `json:"index"`
}

type SessionStorage struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type Config struct {
	Env            string         `json:"env"`
	HTTPServer     HTTPServer     `json:"http_server"`
	Postgres       Postgres       `json:"postgres"`
	Meilisearch    Meilisearch    `json:"meilisearch"`
	SessionStorage SessionStorage `json:"session_storage"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")

	var cfg Config

	if configPath == "" {
		log.Fatalf("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	fin, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %s", configPath)
	}
	defer func(fin *os.File) {
		err := fin.Close()
		if err != nil {
			log.Printf("Failed to close config file")
		}
	}(fin)

	if err := json.NewDecoder(fin).Decode(&cfg); err != nil {
		log.Fatalf("cannot parse config file: %s", err)
	}

	cfg.Postgres.URL = os.ExpandEnv(cfg.Postgres.URL)
	cfg.SessionStorage.Password = os.ExpandEnv(cfg.SessionStorage.Password)
	cfg.SessionStorage.Address = os.ExpandEnv(cfg.SessionStorage.Address)

	cfg.Meilisearch.ApiKey = os.ExpandEnv(cfg.Meilisearch.ApiKey)
	cfg.Meilisearch.URL = os.ExpandEnv(cfg.Meilisearch.URL)

	return &cfg
}
