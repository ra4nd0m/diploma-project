package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"user-service/internal/dto"
	"user-service/internal/middleware"
	"user-service/internal/models"

	"github.com/google/uuid"
)

type UserService interface {
	GetOrCreateUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateUserPreferences(ctx context.Context, id uuid.UUID, preferences json.RawMessage) error
}

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMeContext(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetOrCreateUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	resp := dto.UserResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Preferences: user.Preferences,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Preferences are a stub for this MVP, if the frontend needs to store some user-specific data, it can be stored here as a JSON blob
