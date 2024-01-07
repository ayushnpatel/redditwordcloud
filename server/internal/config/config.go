package config

import (
	"log"
	"path"
	"path/filepath"
	"redditwordcloud/internal/mongodb"
	"redditwordcloud/internal/newrelic"
	"redditwordcloud/internal/reddit"
	"runtime"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type EnvironmentConfig struct {
	LogLevel string
	APIKey   string
}

type Config struct {
	LogLevel       string `env:"LOG_LEVEL,required"`
	Env            string `env:"ENV,required"`
	MongoDBConfig  mongodb.MongoDBConfig
	RedditConfig   reddit.RedditConfig     `envPrefix:"REDDIT_"`
	NewRelicConfig newrelic.NewRelicConfig `envPrefix:"NEW_RELIC_"`
}

const (
	Local      = "LOCAL"
	Dev        = "DEV"
	Production = "PRODUCTION"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(path.Join(basepath, "../../.env")); err != nil {
		log.Print("No .env file found")
	}
}

// New returns a new Config struct
func Load() *Config {

	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}
