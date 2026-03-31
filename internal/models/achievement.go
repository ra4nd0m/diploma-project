package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Achievement struct {
	ID               int64
	Name             string
	Description      string
	IconLink         string
	OwnerID          uuid.UUID
	CohortID         int64
	AccessModeID     int64
	IssuanceKindID   int64
	ConditionTypeID  int64
	ConditionPayload json.RawMessage
}
