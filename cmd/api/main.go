package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/internal/config"
	"user-service/internal/handlers"
	"user-service/internal/logx"
	"user-service/internal/repo"
	"user-service/internal/routes"
	"user-service/internal/security"
	"user-service/internal/services"
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

	userRepo := repo.NewUserRepo(db)
	cohortRepo := repo.NewCohortRepo(db)

	userService := services.NewUserService(userRepo)
	inviteTokenManager := security.NewInviteTokenManager(cfg.JWTSecret, cfg.JWTTTL, "user-service")
	accessTokenManager := security.NewAccessTokenManager(cfg.JWTSecret, cfg.JWTTTL, "user-service")
	cohortService := services.NewCohortService(cohortRepo, inviteTokenManager, userService)

	userHandler := handlers.NewUserHandler(userService)
	cohortHandler := handlers.NewCohortHandler(cohortService, inviteTokenManager)

	router := routes.NewRouter(logger, accessTokenManager, cohortHandler, userHandler)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		logger.Info("starting http server", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	shutdownSignalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrCh:
		if err != nil {
			logger.Error("serve http", "error", slog.Any("error", err))
			os.Exit(1)
		}
	case <-shutdownSignalCtx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", slog.Any("error", err))
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("force close server failed", "error", slog.Any("error", closeErr))
		}
		os.Exit(1)
	}

	logger.Info("http server stopped")

}
