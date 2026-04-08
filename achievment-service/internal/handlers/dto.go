package handlers

import (
	achievement_creation "achievement-service/internal/services/achievement_creation"
	achievementissue "achievement-service/internal/services/achievement_issue"
	achievement_reading_service "achievement-service/internal/services/achievement_reading"
	"encoding/json"

	"github.com/google/uuid"
)

type rawJSON []byte

func (r rawJSON) MarshalJSON() ([]byte, error) {
	return json.RawMessage(r).MarshalJSON()
}

func (r *rawJSON) UnmarshalJSON(data []byte) error {
	var m json.RawMessage
	if err := m.UnmarshalJSON(data); err != nil {
		return err
	}
	*r = rawJSON(m)
	return nil
}

type createAchievementRequestDTO struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	IconLink         string  `json:"icon_link"`
	CohortID         int64   `json:"cohort_id"`
	ConditionType    *string `json:"condition_type,omitempty"`
	IssuanceKind     string  `json:"issuance_kind"`
	ConditionPayload rawJSON `json:"condition_payload,omitempty" swaggertype:"object"`
}

func (d createAchievementRequestDTO) toInput(ownerID uuid.UUID) achievement_creation.Input {
	return achievement_creation.Input{
		Name:             d.Name,
		Description:      d.Description,
		IconLink:         d.IconLink,
		CohortID:         d.CohortID,
		OwnerID:          ownerID,
		ConditionType:    d.ConditionType,
		IssuanceKind:     d.IssuanceKind,
		ConditionPayload: json.RawMessage(d.ConditionPayload),
	}
}

type createAchievementResponseDTO struct {
	ID int64 `json:"id"`
}

type errorResponseDTO struct {
	Error string `json:"error"`
}

type issueAchievementRequestDTO struct {
	AchievementID    int64   `json:"achievement_id"`
	RecipientID      string  `json:"recipient_id"`
	AdditionalDetail *string `json:"additional_detail,omitempty"`
}

func (d issueAchievementRequestDTO) toInput(issuerID uuid.UUID) (achievementissue.Input, error) {
	recipientID, err := uuid.Parse(d.RecipientID)
	if err != nil {
		return achievementissue.Input{}, err
	}

	return achievementissue.Input{
		AchievementID:    d.AchievementID,
		RecipientID:      recipientID,
		IssuerID:         issuerID,
		AdditionalDetail: d.AdditionalDetail,
	}, nil
}

type issueAchievementResponseDTO struct {
	ID int64 `json:"id"`
}

func issueAchievementResponseFromOutput(out *achievementissue.Output) issueAchievementResponseDTO {
	if out == nil {
		return issueAchievementResponseDTO{}
	}
	return issueAchievementResponseDTO{ID: out.ID}
}

type lookupValueDTO struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type achievementResponseDTO struct {
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	IconLink         string          `json:"icon_link"`
	CohortID         int64           `json:"cohort_id"`
	OwnerID          string          `json:"owner_id"`
	AccessMode       lookupValueDTO  `json:"access_mode"`
	IssuanceKind     lookupValueDTO  `json:"issuance_kind"`
	ConditionType    *lookupValueDTO `json:"condition_type,omitempty"`
	ConditionPayload rawJSON         `json:"condition_payload,omitempty" swaggertype:"object"`
	IssuanceID       *int64          `json:"issuance_id,omitempty"`
	Status           *statusDTO      `json:"status,omitempty"`
	AdditionalDetail *string         `json:"additional_detail,omitempty"`
	ProgressPayload  rawJSON         `json:"progress_payload,omitempty" swaggertype:"object"`
}

type statusDTO struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
}

func achievementResponseFromOutput(out *achievement_reading_service.Output) *achievementResponseDTO {
	if out == nil {
		return nil
	}

	var conditionType *lookupValueDTO
	if out.ConditionType != nil {
		conditionType = &lookupValueDTO{
			ID:   out.ConditionType.ID,
			Code: out.ConditionType.Code,
			Name: out.ConditionType.Name,
		}
	}

	return &achievementResponseDTO{
		ID:          out.ID,
		Name:        out.Name,
		Description: out.Description,
		IconLink:    out.IconLink,
		CohortID:    out.CohortID,
		OwnerID:     out.OwnerID.String(),
		AccessMode: lookupValueDTO{
			ID:   out.AccessMode.ID,
			Code: out.AccessMode.Code,
			Name: out.AccessMode.Name,
		},
		IssuanceKind: lookupValueDTO{
			ID:   out.IssuanceKind.ID,
			Code: out.IssuanceKind.Code,
			Name: out.IssuanceKind.Name,
		},
		ConditionType:    conditionType,
		ConditionPayload: rawJSON(out.ConditionPayload),
		IssuanceID:       out.IssuanceID,
		Status:           statusFromOutput(out.Status),
		AdditionalDetail: out.AdditionalDetail,
		ProgressPayload:  rawJSON(out.ProgressPayload),
	}
}

func statusFromOutput(status *achievement_reading_service.AchievementStatus) *statusDTO {
	if status == nil {
		return nil
	}

	return &statusDTO{
		ID:   status.ID,
		Code: status.Code,
	}
}

func achievementsResponseFromOutput(items []*achievement_reading_service.Output) []*achievementResponseDTO {
	result := make([]*achievementResponseDTO, 0, len(items))
	for _, item := range items {
		result = append(result, achievementResponseFromOutput(item))
	}
	return result
}
