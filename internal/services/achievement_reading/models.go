package achievement_reading_service

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Output struct {
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	IconLink         string          `json:"icon_link"`
	CohortID         int64           `json:"cohort_id"`
	OwnerID          uuid.UUID       `json:"owner_id"`
	AccessMode       LookupValue     `json:"access_mode"`
	IssuanceKind     LookupValue     `json:"issuance_kind"`
	ConditionType    *LookupValue    `json:"condition_type,omitempty"`
	ConditionPayload json.RawMessage `json:"condition_payload,omitempty"`
}

type LookupValue struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
