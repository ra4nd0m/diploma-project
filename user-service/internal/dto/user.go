package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// UserResponse represents a user's profile information.
type UserResponse struct {
	ID          uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	DisplayName string          `json:"display_name" example:"John Doe"`
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
}
