package routes

import (
	"log/slog"
	"net/http"
	"user-service/internal/handlers"
	authmiddleware "user-service/internal/middleware"

	_ "user-service/docs"

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

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	auth := authmiddleware.NewAuthMiddleware(authManager)
	internalTokenValidator := authmiddleware.NewInternalTokenMiddleware(internalToken)

	r.Group(func(protected chi.Router) {
		protected.Use(auth)

		protected.Get("/me", userHandler.GetMeContext)

		protected.Get("/cohorts", cohortHandler.GetCohorts)
		protected.Post("/cohorts", cohortHandler.CreateCohort)
		protected.Get("/cohorts/{id}/members", cohortHandler.GetCohortMembers)
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
