package routes

import (
	"log/slog"
	"net/http"
	"user-service/docs"
	_ "user-service/docs"
	"user-service/internal/handlers"
	authmiddleware "user-service/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(
	logger *slog.Logger,
	authManager authmiddleware.AccessTokenManager,
	cohortHandler *handlers.CohortHandler,
	userHandler *handlers.UserHandler,
	internalToken string,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(authmiddleware.NewRequestLogMiddleware(logger))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/swagger/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(docs.SwaggerInfo.ReadDoc())); err != nil {
			logger.Error("failed to write swagger doc", "error", err)
		}
	})

	auth := authmiddleware.NewAuthMiddleware(authManager)
	internalTokenValidator := authmiddleware.NewInternalTokenMiddleware(internalToken)

	r.Group(func(protected chi.Router) {
		protected.Use(auth)

		protected.Get("/me", userHandler.GetMeContext)

		protected.Get("/cohorts", cohortHandler.GetCohorts)
		protected.Post("/cohorts", cohortHandler.CreateCohort)
		protected.Get("/cohorts/{id}/members", cohortHandler.GetCohortMembers)
		protected.Delete("/cohorts/{id}/members/{user_id}", cohortHandler.RemoveUserFromCohort)
		protected.Post("/cohorts/join", cohortHandler.JoinCohort)
		protected.Post("/cohorts/{id}/invite", cohortHandler.GenerateInviteToken)
	})

	r.Group(func(internal chi.Router) {
		internal.Use(internalTokenValidator)
		internal.Post("/internal/cohorts/can-edit", cohortHandler.IsOwner)
		internal.Post("/internal/cohorts/is-user-in", cohortHandler.IsUserIn)
	})

	return r
}
