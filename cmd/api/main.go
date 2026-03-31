package main

import (
	"achievement-service/internal/config"
	"achievement-service/internal/logx"
	"achievement-service/internal/repo"
	"log"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config %v", err)
	}

	logger := logx.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(logger)

	db, err := repo.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("close database", "error", slog.Any("error", err))
		}
	}()

	if err := repo.RunMigrations(db, "migrations"); err != nil {
		logger.Error("run migrations", "error", slog.Any("error", err))
		os.Exit(1)
	}
}
