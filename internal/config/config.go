package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	CohortURL   string
	InternalToken string
	JWTSecret   string
	JWTTTL      time.Duration
	JWTIssuer   string
	LogLevel    string
	LogFormat   string
}

func Load() (Config, error) {
	ttl, err := time.ParseDuration(getEnv("JWT_TTL", "24h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_TTL %w", err)
	}

	cfg := Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://user:password@localhost/db"),
		CohortURL:   getEnv("COHORT_URL", "http://localhost:8081"),
		InternalToken: getEnv("INTERNAL_TOKEN", "internal-token"),
		JWTSecret:   getEnv("JWT_SECRET", "secret"),
		JWTTTL:      ttl,
		JWTIssuer:   getEnv("JWT_ISSUER", "achievement-service"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
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
