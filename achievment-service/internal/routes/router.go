package routes

import (
	"achievement-service/docs"
	_ "achievement-service/docs"
	"achievement-service/internal/handlers"
	"achievement-service/internal/middleware"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(
	logger *slog.Logger,
	achievementHandler *handlers.AchievementHandler,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.NewRequestLogMiddleware(logger))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(docs.SwaggerInfo.ReadDoc())); err != nil {
			logger.Error("write swagger doc", "error", slog.Any("error", err))
		}
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.MethodFunc(http.MethodGet, "/achievements", achievementHandler.GetAchievements)
		r.MethodFunc(http.MethodGet, "/achievements/owned", achievementHandler.GetOwnedAchievements)
		r.MethodFunc(http.MethodGet, "/achievements/recipient/{recipientID}", achievementHandler.GetRecipientAchievements)
		r.MethodFunc(http.MethodPost, "/achievements", achievementHandler.CreateAchievement)
		r.MethodFunc(http.MethodPost, "/achievements/issue", achievementHandler.IssueAchievement)
	})

	return r
}
