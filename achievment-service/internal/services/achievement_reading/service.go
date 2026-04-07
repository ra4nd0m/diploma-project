package achievement_reading_service

import (
	"achievement-service/internal/models"
	"achievement-service/internal/services"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type AchievementReadingRepo interface {
	ListVisibleAchievements(
		ctx context.Context,
		userID uuid.UUID,
		cohortIDs []int64,
		publicAccessModeID int64,
		cohortAccessModeID int64,
		privateAccessModeID int64,
	) ([]*models.Achievement, error)
	GetAchievementsByOwner(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]*models.Achievement, error)
	ListAchievementsForRecipient(
		ctx context.Context,
		requestUserID uuid.UUID,
		recipientID uuid.UUID,
		cohortIDs []int64,
		publicAccessModeID int64,
		cohortAccessModeID int64,
		privateAccessModeID int64,
	) ([]*models.PersonalAchievement, error)
}

type AchievementReadingLookupRepo interface {
	GetAccessModeByCode(ctx context.Context, code string) (*models.AccessMode, error)
	GetAccessModeByID(ctx context.Context, id int64) (*models.AccessMode, error)
	GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error)
	GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error)
}

type AchievementReadingService struct {
	repo       AchievementReadingRepo
	lookupRepo AchievementReadingLookupRepo
	authz      authz
}

type authz interface {
	RequireUserInCohorts(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]int64, error)
}

func NewAchievementReadingService(repo AchievementReadingRepo, lookupRepo AchievementReadingLookupRepo, authz authz) *AchievementReadingService {
	return &AchievementReadingService{repo: repo, lookupRepo: lookupRepo, authz: authz}
}

func (s *AchievementReadingService) GetVisibleAchievements(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]*Output, error) {
	if userID == uuid.Nil {
		return nil, services.ErrInvalidInput
	}

	validatedCohortIDs, err := s.authz.RequireUserInCohorts(ctx, userID, cohortIDs)
	if err != nil {
		return nil, err
	}

	publicModeID, cohortModeID, privateModeID, err := s.getVisibilityAccessModeIDs(ctx)
	if err != nil {
		return nil, err
	}

	achievements, err := s.repo.ListVisibleAchievements(ctx, userID, validatedCohortIDs, publicModeID, cohortModeID, privateModeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get visible achievements: %w", err)
	}

	return s.assembleAndVerifyAchievements(ctx, achievements)
}

func (s *AchievementReadingService) GetOwnedAchievements(ctx context.Context, ownerID uuid.UUID, cohortIDs []int64) ([]*Output, error) {
	if ownerID == uuid.Nil {
		return nil, services.ErrInvalidInput
	}

	validatedCohortIDs, err := s.authz.RequireUserInCohorts(ctx, ownerID, cohortIDs)
	if err != nil {
		return nil, err
	}

	achievements, err := s.repo.GetAchievementsByOwner(ctx, ownerID, validatedCohortIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get owner achievements: %w", err)
	}

	return s.assembleAndVerifyAchievements(ctx, achievements)
}

func (s *AchievementReadingService) GetRecipientAchievements(ctx context.Context, requestUserID, recipientID uuid.UUID, cohortIDs []int64) ([]*Output, error) {
	if requestUserID == uuid.Nil || recipientID == uuid.Nil {
		return nil, services.ErrInvalidInput
	}

	validatedCohortIDs, err := s.authz.RequireUserInCohorts(ctx, requestUserID, cohortIDs)
	if err != nil {
		return nil, err
	}

	publicModeID, cohortModeID, privateModeID, err := s.getVisibilityAccessModeIDs(ctx)
	if err != nil {
		return nil, err
	}

	achievements, err := s.repo.ListAchievementsForRecipient(ctx, requestUserID, recipientID, validatedCohortIDs, publicModeID, cohortModeID, privateModeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipient achievements: %w", err)
	}

	return s.assembleAndVerifyPersonalAchievements(ctx, achievements)
}

func (s *AchievementReadingService) getVisibilityAccessModeIDs(ctx context.Context) (int64, int64, int64, error) {
	publicMode, err := s.lookupRepo.GetAccessModeByCode(ctx, models.AccessModePublic)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to resolve public access mode: %w", err)
	}
	if publicMode == nil {
		return 0, 0, 0, services.ErrAccessModeNotFound
	}

	cohortMode, err := s.lookupRepo.GetAccessModeByCode(ctx, models.AccessModeCohort)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to resolve cohort access mode: %w", err)
	}
	if cohortMode == nil {
		return 0, 0, 0, services.ErrAccessModeNotFound
	}

	privateMode, err := s.lookupRepo.GetAccessModeByCode(ctx, models.AccessModePrivate)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to resolve private access mode: %w", err)
	}
	if privateMode == nil {
		return 0, 0, 0, services.ErrAccessModeNotFound
	}

	return publicMode.ID, cohortMode.ID, privateMode.ID, nil
}

func (s *AchievementReadingService) assembleAndVerifyAchievements(ctx context.Context, achievements []*models.Achievement) ([]*Output, error) {

	outputs := make([]*Output, 0, len(achievements))

	for _, achievement := range achievements {
		if achievement == nil {
			continue
		}

		accessMode, err := s.lookupRepo.GetAccessModeByID(ctx, achievement.AccessModeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get access mode: %w", err)
		}
		if accessMode == nil {
			return nil, services.ErrAccessModeNotFound
		}

		issuanceKind, err := s.lookupRepo.GetIssuanceKindByID(ctx, achievement.IssuanceKindID)
		if err != nil {
			return nil, fmt.Errorf("failed to get issuance kind: %w", err)
		}
		if issuanceKind == nil {
			return nil, services.ErrIssuanceKindNotFound
		}

		var conditionTypeOutput *LookupValue

		if achievement.ConditionTypeID > 0 {
			conditionType, err := s.lookupRepo.GetConditionTypeByID(ctx, achievement.ConditionTypeID)
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

func (s *AchievementReadingService) assembleAndVerifyPersonalAchievements(ctx context.Context, achievements []*models.PersonalAchievement) ([]*Output, error) {
	outputs := make([]*Output, 0, len(achievements))

	for _, achievement := range achievements {
		if achievement == nil {
			continue
		}

		accessMode, err := s.lookupRepo.GetAccessModeByID(ctx, achievement.AccessModeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get access mode: %w", err)
		}
		if accessMode == nil {
			return nil, services.ErrAccessModeNotFound
		}

		issuanceKind, err := s.lookupRepo.GetIssuanceKindByID(ctx, achievement.IssuanceKindID)
		if err != nil {
			return nil, fmt.Errorf("failed to get issuance kind: %w", err)
		}
		if issuanceKind == nil {
			return nil, services.ErrIssuanceKindNotFound
		}

		var conditionTypeOutput *LookupValue
		if achievement.ConditionTypeID > 0 {
			conditionType, err := s.lookupRepo.GetConditionTypeByID(ctx, achievement.ConditionTypeID)
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

		var status *AchievementStatus
		if achievement.StatusID != nil && achievement.StatusCode != nil {
			status = &AchievementStatus{
				ID:   *achievement.StatusID,
				Code: *achievement.StatusCode,
			}
		}

		outputs = append(outputs, &Output{
			ID:          achievement.AchievementID,
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
			IssuanceID:       achievement.IssuanceID,
			Status:           status,
			AdditionalDetail: achievement.AdditionalDetail,
			ProgressPayload:  achievement.ProgressPayload,
		})
	}

	return outputs, nil
}
