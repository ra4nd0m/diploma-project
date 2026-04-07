package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type PersonalAchievement struct {
	AchievementID    int64
	Name             string
	Description      string
	IconLink         string
	OwnerID          uuid.UUID
	CohortID         int64
	AccessModeID     int64
	IssuanceKindID   int64
	ConditionTypeID  int64
	ConditionPayload json.RawMessage

	IssuanceID       *int64
	StatusID         *int64
	StatusCode       *string
	AdditionalDetail *string
	ProgressPayload  json.RawMessage
}
