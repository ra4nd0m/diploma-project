package main

import (
	cohortclient "achievement-service/internal/clients/cohort"
	"achievement-service/internal/config"
	"achievement-service/internal/handlers"
	"achievement-service/internal/logx"
	"achievement-service/internal/middleware"
	"achievement-service/internal/repo"
	"achievement-service/internal/routes"
	"achievement-service/internal/security"
	"achievement-service/internal/services/achievement_creation"
	achievementissue "achievement-service/internal/services/achievement_issue"
	achievement_reading_service "achievement-service/internal/services/achievement_reading"
	"achievement-service/internal/services/authz"
	"log"
	"log/slog"
	"net/http"
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

	achievementRepo := repo.NewAchievementRepo(db)
	issuanceLifecycleRepo := repo.NewIssuanceLifecycleRepo(db)
	lookupRepo := repo.NewLookupRepo(db)

	tokenManager := security.NewAccessTokenManager(cfg.JWTSecret, cfg.JWTTTL, cfg.JWTIssuer)

	cohortClient := cohortclient.NewClient(cfg.CohortURL, cfg.InternalToken)

	authzService := authz.NewService(cohortClient)

	achievementCreationService := achievement_creation.NewAchievementCreationService(achievementRepo, lookupRepo, authzService)
	achievementIssueService := achievementissue.NewService(achievementRepo, lookupRepo, issuanceLifecycleRepo, authzService)
	achievementReadingService := achievement_reading_service.NewAchievementReadingService(achievementRepo, lookupRepo)

	achievementHandler := handlers.NewAchievementHandler(
		achievementCreationService,
		achievementIssueService,
		achievementReadingService,
	)

	authMiddleware := middleware.NewAuthMiddleware(tokenManager)
	router := routes.NewRouter(logger, achievementHandler, authMiddleware)

	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}

	logger.Info("starting http server", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("http server stopped", "error", slog.Any("error", err))
		os.Exit(1)
	}

}
