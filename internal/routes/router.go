package routes

import (
	"log/slog"
	"net/http"
	"time"
	"user-service/internal/handlers"
	authmiddleware "user-service/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	r.Use(requestLogger(logger))

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
	})

	return r
}

type statusCapturingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if logger == nil {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			wrapped := &statusCapturingResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration", time.Since(start).String(),
			)
		})
	}
}
