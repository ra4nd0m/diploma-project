package routes

import (
	"achievement-service/internal/handlers"
	"achievement-service/internal/middleware"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	logger *slog.Logger,
	achievementHandler *handlers.AchievementHandler,
	authMiddleware func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.NewRequestLogMiddleware(logger))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
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
