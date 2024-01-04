package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvironmentConfig struct {
	LogLevel string
	APIKey   string
}

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	Env      string `env:"ENV"`
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
	fmt.Println(path.Join(basepath, "../../.env"))
	if err := godotenv.Load(path.Join(basepath, "../../.env")); err != nil {
		log.Print("No .env file found")
	}
}

// New returns a new Config struct
func Load() *Config {
	return &Config{
		LogLevel: GetEnv("LOG_LEVEL", "Debug"),
		Env:      GetEnv("ENV", "LOCAL"),
	}
}

// Simple helper function to read an environment or return a default value
func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// Simple helper function to read an environment variable into integer or return a default value
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := GetEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

// Helper to read an environment variable into a bool or return default value
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}
