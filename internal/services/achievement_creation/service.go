package achievement_creation

import (
	"achievement-service/internal/models"
	"achievement-service/internal/services"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type AchievementCreationRepo interface {
	GetAccessModeByCode(ctx context.Context, code string) (*models.AccessMode, error)
	GetIssuanceKindByCode(ctx context.Context, code string) (*models.IssuanceKind, error)
	GetConditionTypeByCode(ctx context.Context, code string) (*models.ConditionType, error)

	CreateAchievement(ctx context.Context, achievement models.Achievement) (int64, error)
}

type AchievementCreationService struct {
	repo AchievementCreationRepo
}

func NewAchievementCreationService(repo AchievementCreationRepo) *AchievementCreationService {
	return &AchievementCreationService{repo: repo}
}

func (s *AchievementCreationService) CreateAchievement(ctx context.Context, input Input) (int64, error) {
	if input.Name == "" {
		return 0, services.ErrInvalidInput
	}
	if input.CohortID <= 0 {
		return 0, services.ErrInvalidInput
	}
	if input.OwnerID == uuid.Nil {
		return 0, services.ErrInvalidInput
	}
	if input.IssuanceKind == "" {
		return 0, services.ErrInvalidInput
	}

	accessMode, err := s.repo.GetAccessModeByCode(ctx, models.AccessModeCohort)
	if err != nil {
		return 0, fmt.Errorf("access mode by code: %w", err)
	}

	issuanceKind, err := s.repo.GetIssuanceKindByCode(ctx, input.IssuanceKind)
	if err != nil {
		return 0, fmt.Errorf("issuance kind by code: %w", err)
	}

	var conditionTypeID int64
	var conditionPayload json.RawMessage

	if input.ConditionType != nil {
		conditionType, err := s.repo.GetConditionTypeByCode(ctx, *input.ConditionType)
		if err != nil {
			return 0, fmt.Errorf("condition type by code: %w", err)
		}
		if conditionType == nil {
			return 0, services.ErrConditionTypeNotFound
		}

		validatedPayload, err := validateCondition(*input.ConditionType, input.ConditionPayload)

		if err != nil {
			return 0, fmt.Errorf("validate condition: %w", err)
		}

		conditionPayload = validatedPayload
		conditionTypeID = conditionType.ID
	} else if len(input.ConditionPayload) > 0 && string(input.ConditionPayload) != "null" {
		return 0, services.ErrInvalidCondition
	}

	return s.repo.CreateAchievement(ctx, models.Achievement{
		Name:             input.Name,
		CohortID:         input.CohortID,
		OwnerID:          input.OwnerID,
		AccessModeID:     accessMode.ID,
		IssuanceKindID:   issuanceKind.ID,
		ConditionTypeID:  conditionTypeID,
		ConditionPayload: conditionPayload,
	})

}

func validateCondition(conditionType string, raw json.RawMessage) (json.RawMessage, error) {
	switch conditionType {
	case models.ConditionTypeAllOf:
		var payload AllOfConditionPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return nil, services.ErrInvalidCondition
		}
		if len(payload.AchievementIDs) == 0 {
			return nil, services.ErrInvalidCondition
		}
		for _, id := range payload.AchievementIDs {
			if id <= 0 {
				return nil, services.ErrInvalidCondition
			}
		}
		return json.Marshal(payload)
	default:
		return nil, services.ErrConditionTypeNotFound
	}
}
