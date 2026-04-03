package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"user-service/internal/dto"
	"user-service/internal/security"
)

type AccessTokenManager interface {
	ParseAccessToken(tokenStr string) (*security.AccessClaims, error)
}

type contextKey string

const accessClaimsKey contextKey = "accessClaims"

func NewAuthMiddleware(manager AccessTokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			const prefix = "Bearer "

			if !strings.HasPrefix(authHeader, prefix) {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if tokenStr == "" {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := manager.ParseAccessToken(tokenStr)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), accessClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*security.AccessClaims, bool) {
	claims, ok := ctx.Value(accessClaimsKey).(*security.AccessClaims)
	return claims, ok
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Errors: message,
	})
}

func NewInternalTokenMiddleware(expectedToken string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Internal-Token")
			if token == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing X-Internal-Token header")
				return
			}

			if token != expectedToken {
				writeJSONError(w, http.StatusUnauthorized, "invalid X-Internal-Token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
