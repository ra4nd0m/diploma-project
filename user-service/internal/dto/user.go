package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID          uuid.UUID       `json:"id"`
	DisplayName string          `json:"display_name"`
	Preferences json.RawMessage `json:"preferences"`
}
