package config

import (
	"os"
)

type Config struct {
	DBUrl         string
	RedisAddr     string
	Port          string
	RateLimitSpec string
	OTELCollector string
	APIKey        string
	RMAPI         string
}

func Load() *Config {
	return &Config{
		DBUrl:         getenv("DATABASE_URL", "postgres://postgres:password@db:5432/myapp?sslmode=disable"),
		RedisAddr:     getenv("REDIS_ADDR", "redis:6379"),
		Port:          getenv("PORT", "8080"),
		RateLimitSpec: getenv("RATE_LIMIT_SPEC", "100-M"),
		OTELCollector: getenv("OTEL_COLLECTOR_URL", "http://otel-collector:4317"),
		APIKey:        getenv("API_KEY", "my-secret-key"),
		RMAPI:         getenv("RM_API_ENDPOINT", "https://rickandmortyapi.com/api/character"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
