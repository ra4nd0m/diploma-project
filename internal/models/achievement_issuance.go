package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type AchievementIssuance struct {
	ID               int64
	AchievementID    int64
	RecipientID      uuid.UUID
	IssuerID         uuid.UUID
	StatusID         int64
	AdditionalDetail *string
	ProgressPayload  json.RawMessage
}
