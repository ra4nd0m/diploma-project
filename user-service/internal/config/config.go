package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr      string
	DatabaseURL   string
	JWTSecret     string
	JWTTTL        time.Duration
	LogLevel      string
	LogFormat     string
	InternalToken string
}

func Load() (Config, error) {
	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_TTL %w", err)
	}

	cfg := Config{
		HTTPAddr:      getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://user:password@localhost/db"),
		JWTSecret:     getEnv("JWT_SECRET", "secret"),
		JWTTTL:        ttl,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogFormat:     getEnv("LOG_FORMAT", "text"),
		InternalToken: getEnv("INTERNAL_TOKEN", ""),
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
