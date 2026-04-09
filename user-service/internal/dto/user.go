// Package dto contains data transfer objects for the User Service API.
//
// DTOs define the structure of request and response payloads for API endpoints.
// They handle serialization/deserialization with JSON and provide type safety
// for HTTP communication.
package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// UserResponse represents a user's public profile information.
//
// swagger:model
type UserResponse struct {
	// User's unique identifier (UUID)
	ID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// User's display name
	DisplayName string `json:"display_name" example:"John Doe"`
	// User's preferences stored as arbitrary JSON data
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
}
