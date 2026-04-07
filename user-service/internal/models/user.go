package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	DisplayName string
	Preferences json.RawMessage
}
