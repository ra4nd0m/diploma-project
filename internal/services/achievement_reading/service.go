package achievement_reading_service

import (
	"achievement-service/internal/models"
	"achievement-service/internal/services"
	"context"
	"fmt"
)

type achievementReadingRepo interface {
	GetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error)
	GetAchievements(ctx context.Context, userID int64) ([]*models.Achievement, error)
	GetAccessModeByID(ctx context.Context, id int64) (*models.AccessMode, error)
	GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error)
	GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error)
}

type AchievementReadingService struct {
	repo achievementReadingRepo
}

func NewAchievementReadingService(repo achievementReadingRepo) *AchievementReadingService {
	return &AchievementReadingService{repo: repo}
}

func (s *AchievementReadingService) GetAchievement(ctx context.Context, achievementID int64) (*Output, error) {
	if achievementID <= 0 {
		return nil, services.ErrInvalidInput
	}
	achievement, err := s.repo.GetAchievement(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}
	if achievement == nil {
		return nil, services.ErrNotFound
	}

	outputs, err := s.assembleAndVerifyAchievements(ctx, []*models.Achievement{achievement})
	if err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, services.ErrNotFound
	}

	return outputs[0], nil
}

func (s *AchievementReadingService) GetAchievements(ctx context.Context, userID int64) ([]*Output, error) {
	if userID <= 0 {
		return nil, services.ErrInvalidInput
	}

	achievements, err := s.repo.GetAchievements(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	return s.assembleAndVerifyAchievements(ctx, achievements)
}

func (s *AchievementReadingService) assembleAndVerifyAchievements(ctx context.Context, achievements []*models.Achievement) ([]*Output, error) {

	outputs := make([]*Output, 0, len(achievements))

	for _, achievement := range achievements {
		if achievement == nil {
			continue
		}

		accessMode, err := s.repo.GetAccessModeByID(ctx, achievement.AccessModeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get access mode: %w", err)
		}
		if accessMode == nil {
			return nil, services.ErrAccessModeNotFound
		}

		issuanceKind, err := s.repo.GetIssuanceKindByID(ctx, achievement.IssuanceKindID)
		if err != nil {
			return nil, fmt.Errorf("failed to get issuance kind: %w", err)
		}
		if issuanceKind == nil {
			return nil, services.ErrIssuanceKindNotFound
		}

		var conditionTypeOutput *LookupValue

		if achievement.ConditionTypeID > 0 {
			conditionType, err := s.repo.GetConditionTypeByID(ctx, achievement.ConditionTypeID)
			if err != nil {
				return nil, fmt.Errorf("failed to get condition type: %w", err)
			}
			if conditionType == nil {
				return nil, services.ErrConditionTypeNotFound
			}
			conditionTypeOutput = &LookupValue{
				ID:   conditionType.ID,
				Code: conditionType.Code,
				Name: conditionType.Name,
			}
		}

		outputs = append(outputs, &Output{
			ID:          achievement.ID,
			Name:        achievement.Name,
			Description: achievement.Description,
			IconLink:    achievement.IconLink,
			CohortID:    achievement.CohortID,
			OwnerID:     achievement.OwnerID,
			AccessMode: LookupValue{
				ID:   accessMode.ID,
				Code: accessMode.Code,
				Name: accessMode.Name,
			},
			IssuanceKind: LookupValue{
				ID:   issuanceKind.ID,
				Code: issuanceKind.Code,
				Name: issuanceKind.Name,
			},
			ConditionType:    conditionTypeOutput,
			ConditionPayload: achievement.ConditionPayload,
		})
	}

	return outputs, nil
}
