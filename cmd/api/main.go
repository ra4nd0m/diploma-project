package main

import (
	"log"
	"log/slog"
	"os"
	"user-service/internal/config"
	"user-service/internal/logx"
	"user-service/internal/repo"
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
