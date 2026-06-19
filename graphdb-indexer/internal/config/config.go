package config

import "os"

type HTTPServer struct {
	Address string
}

type MessageBroker struct {
	URL string
}

type GraphDB struct {
	URI  string
	User string
	Pass string
}

type FeatureStorage struct {
	Addr string
	Pass string
}

type Config struct {
	Env string

	HTTPServer     HTTPServer
	MessageBroker  MessageBroker
	GraphDB        GraphDB
	FeatureStorage FeatureStorage
}

func MustLoad() *Config {
	return &Config{
		Env: "local", // TODO: load from env

		HTTPServer: HTTPServer{
			Address: "0.0.0.0:" +
				mustLoadENV("GRAPHDB_INDEXER_PORT"),
		},

		MessageBroker: MessageBroker{
			URL: "nats://" +
				mustLoadENV("MESSAGE_BROKER_HOST") + ":" +
				mustLoadENV("MESSAGE_BROKER_PORT"),
		},

		GraphDB: GraphDB{
			URI: "bolt://" +
				mustLoadENV("GRAPHDB_HOST") + ":" +
				mustLoadENV("GRAPHDB_PORT"),
			User: "neo4j",
			Pass: mustLoadENV("GRAPHDB_PASSWORD"),
		},

		FeatureStorage: FeatureStorage{
			Addr: mustLoadENV("FEATURE_STORAGE_HOST") + ":" +
				mustLoadENV("FEATURE_STORAGE_PORT"),
			Pass: mustLoadENV("FEATURE_STORAGE_PASSWORD"),
		},
	}
}

func mustLoadENV(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic(key + " ENV var is not set")
}
