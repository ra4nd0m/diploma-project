package handlers

import (
	achievement_creation "achievement-service/internal/services/achievement_creation"
	achievementissue "achievement-service/internal/services/achievement_issue"
	achievement_reading_service "achievement-service/internal/services/achievement_reading"
	"encoding/json"

	"github.com/google/uuid"
)

// createAchievementRequestDTO represents the request payload for creating a new achievement
type createAchievementRequestDTO struct {
	// Name of the achievement
	Name string `json:"name"`
	// Description explaining what the achievement is about
	Description string `json:"description"`
	// IconLink URL pointing to the achievement icon/image
	IconLink string `json:"icon_link"`
	// CohortID the ID of the cohort this achievement belongs to
	CohortID int64 `json:"cohort_id"`
	// ConditionType optional type of condition for automatic issuance
	ConditionType *string `json:"condition_type,omitempty"`
	// IssuanceKind type of issuance (e.g., 'manual', 'automatic')
	IssuanceKind string `json:"issuance_kind"`
	// ConditionPayload optional JSON payload containing condition details
	ConditionPayload json.RawMessage `json:"condition_payload,omitempty"`
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
		ConditionPayload: d.ConditionPayload,
	}
}

// createAchievementResponseDTO contains the response after creating an achievement
type createAchievementResponseDTO struct {
	// ID of the newly created achievement
	ID int64 `json:"id"`
}

// issueAchievementRequestDTO represents the request payload for issuing an achievement to a user
type issueAchievementRequestDTO struct {
	// AchievementID the ID of the achievement to issue
	AchievementID int64 `json:"achievement_id"`
	// RecipientID UUID of the user receiving the achievement
	RecipientID string `json:"recipient_id"`
	// AdditionalDetail optional additional information about the issuance
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

// issueAchievementResponseDTO contains the response after issuing an achievement
type issueAchievementResponseDTO struct {
	// ID of the created achievement issuance record
	ID int64 `json:"id"`
}

func issueAchievementResponseFromOutput(out *achievementissue.Output) issueAchievementResponseDTO {
	if out == nil {
		return issueAchievementResponseDTO{}
	}
	return issueAchievementResponseDTO{ID: out.ID}
}

// lookupValueDTO represents a reference to a lookup table entry (e.g., status, type, kind)
type lookupValueDTO struct {
	// ID unique identifier of the lookup value
	ID int64 `json:"id"`
	// Code string code of the lookup value
	Code string `json:"code"`
	// Name human-readable name of the lookup value
	Name string `json:"name"`
}

// achievementResponseDTO contains detailed information about an achievement
type achievementResponseDTO struct {
	// ID unique identifier of the achievement
	ID int64 `json:"id"`
	// Name of the achievement
	Name string `json:"name"`
	// Description explaining what the achievement is about
	Description string `json:"description"`
	// IconLink URL pointing to the achievement icon/image
	IconLink string `json:"icon_link"`
	// CohortID the ID of the cohort this achievement belongs to
	CohortID int64 `json:"cohort_id"`
	// OwnerID UUID of the user who created/owns this achievement
	OwnerID string `json:"owner_id"`
	// AccessMode the access level/mode for this achievement
	AccessMode lookupValueDTO `json:"access_mode"`
	// IssuanceKind type of issuance (e.g., 'manual', 'automatic')
	IssuanceKind lookupValueDTO `json:"issuance_kind"`
	// ConditionType optional type of condition for automatic issuance
	ConditionType *lookupValueDTO `json:"condition_type,omitempty"`
	// ConditionPayload optional JSON payload containing condition details
	ConditionPayload json.RawMessage `json:"condition_payload,omitempty"`
	// IssuanceID optional ID of the issuance record (for personal achievements)
	IssuanceID *int64 `json:"issuance_id,omitempty"`
	// Status optional current status of the achievement issuance
	Status *statusDTO `json:"status,omitempty"`
	// AdditionalDetail optional additional information about the achievement
	AdditionalDetail *string `json:"additional_detail,omitempty"`
	// ProgressPayload optional JSON payload containing progress information
	ProgressPayload json.RawMessage `json:"progress_payload,omitempty"`
}

// statusDTO represents the status of an achievement issuance
type statusDTO struct {
	// ID unique identifier of the status
	ID int64 `json:"id"`
	// Code string code of the status
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
		ConditionPayload: out.ConditionPayload,
		IssuanceID:       out.IssuanceID,
		Status:           statusFromOutput(out.Status),
		AdditionalDetail: out.AdditionalDetail,
		ProgressPayload:  out.ProgressPayload,
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
