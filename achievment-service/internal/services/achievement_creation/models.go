package achievement_creation

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Input struct {
	Name             string
	Description      string
	IconLink         string
	CohortID         int64
	OwnerID          uuid.UUID
	ConditionType    *string
	IssuanceKind     string
	ConditionPayload json.RawMessage
}

type AllOfConditionPayload struct {
	AchievementIDs []int64 `json:"achievement_ids"`
}
