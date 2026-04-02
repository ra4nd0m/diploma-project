package achievementissue

import (
	"achievement-service/internal/models"
	"achievement-service/internal/services"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
)

type repo interface {
	GetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error)
	GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error)
	GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error)
	GetAchievementStatusByCode(ctx context.Context, code string) (*models.AchievementStatus, error)

	GetRecipientIssuance(ctx context.Context, achievementID int64, recipientID uuid.UUID) (*models.AchievementIssuance, error)

	IssueAchievement(ctx context.Context, issuance models.AchievementIssuance) (int64, error)

	FindDependentsByStatus(ctx context.Context, recipientID uuid.UUID, cohortID int64, statusID int64) ([]*models.AchievementIssuance, error)
	FindDependentAchievements(ctx context.Context, dependencyAchievementID int64, cohortID int64) ([]*models.Achievement, error)

	UpdateIssuanceProgress(ctx context.Context, issuanceID int64, statusID int64, progressPayload []byte) error
}

type Service struct {
	repo repo
}

func NewService(repo repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) IssueAchievement(ctx context.Context, input Input) (*Output, error) {
	if input.AchievementID <= 0 || input.RecipientID == uuid.Nil || input.IssuerID == uuid.Nil {
		return nil, services.ErrInvalidInput
	}
	achievement, err := s.mustGetAchievement(ctx, input.AchievementID)
	if err != nil {
		return nil, err
	}
	// Here goes the authz check

	if err := s.ensureNotAlreadyIssued(ctx, achievement.ID, input.RecipientID); err != nil {
		return nil, err
	}
	if err := s.ensureManualIssuance(ctx, achievement.IssuanceKindID); err != nil {
		return nil, err
	}
	issuanceID, err := s.createDirectIssuance(ctx, achievement.ID, input.RecipientID, input.IssuerID, input.AdditionalDetail)
	if err != nil {
		return nil, err
	}

	// Best effort dependent progression
	// Main issuance already happened
	if err := s.syncDependentProgress(ctx, achievement, input.RecipientID, input.IssuerID); err != nil {
		return &Output{ID: issuanceID}, err
	}

	return &Output{ID: issuanceID}, nil
}

func (s *Service) mustGetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error) {
	achievement, err := s.repo.GetAchievement(ctx, achievementID)
	if err != nil {
		return nil, fmt.Errorf("get achievement: %w", err)
	}
	if achievement == nil {
		return nil, services.ErrAchievementNotFound
	}
	return achievement, nil
}
func (s *Service) ensureNotAlreadyIssued(ctx context.Context, achievementID int64, recipientID uuid.UUID) error {
	existing, err := s.repo.GetRecipientIssuance(ctx, achievementID, recipientID)
	if err != nil {
		return fmt.Errorf("get recipient issuance: %w", err)
	}
	if existing != nil {
		return services.ErrAlreadyIssued
	}
	return nil
}

func (s *Service) ensureManualIssuance(ctx context.Context, issuanceKindID int64) error {
	kind, err := s.repo.GetIssuanceKindByID(ctx, issuanceKindID)
	if err != nil {
		return fmt.Errorf("get issuance kind: %w", err)
	}
	if kind == nil {
		return services.ErrIssuanceKindNotFound
	}
	if kind.Code != models.IssuanceKindManual {
		return services.ErrForbidden
	}
	return nil
}

func (s *Service) createDirectIssuance(
	ctx context.Context,
	achievementID int64,
	recipientID uuid.UUID,
	issuerID uuid.UUID,
	additionalDetail *string,
) (int64, error) {
	status, err := s.repo.GetAchievementStatusByCode(ctx, models.AchievementStatusIssued)
	if err != nil {
		return 0, fmt.Errorf("get issued status: %w", err)
	}
	if status == nil {
		return 0, services.ErrStatusNotFound
	}

	id, err := s.repo.IssueAchievement(ctx, models.AchievementIssuance{
		AchievementID:    achievementID,
		RecipientID:      recipientID,
		IssuerID:         issuerID,
		StatusID:         status.ID,
		AdditionalDetail: additionalDetail,
	})
	if err != nil {
		return 0, fmt.Errorf("issue achievement: %w", err)
	}

	return id, nil
}

func (s *Service) syncDependentProgress(
	ctx context.Context,
	issuedAchievement *models.Achievement,
	recipientID uuid.UUID,
	issuerID uuid.UUID,
) error {
	inProgressStatus, err := s.repo.GetAchievementStatusByCode(ctx, models.AchievementStatusInProgress)
	if err != nil {
		return fmt.Errorf("get in-progress status: %w", err)
	}
	if inProgressStatus == nil {
		return services.ErrStatusNotFound
	}

	inProgressRows, err := s.repo.FindDependentsByStatus(ctx, recipientID, issuedAchievement.CohortID, inProgressStatus.ID)
	if err != nil {
		return fmt.Errorf("find in-progress dependents: %w", err)
	}

	if len(inProgressRows) > 0 {
		return s.advanceExistingDependents(ctx, inProgressRows, issuedAchievement.ID)
	}

	return s.bootstrapDependents(ctx, issuedAchievement, recipientID, issuerID, inProgressStatus.ID)
}

func (s *Service) advanceExistingDependents(
	ctx context.Context,
	inProgressRows []*models.AchievementIssuance,
	completedAchievementID int64,
) error {
	for _, row := range inProgressRows {
		var payload InProgressPayload
		if err := json.Unmarshal(row.ProgressPayload, &payload); err != nil {
			return services.ErrInvalidCondition
		}

		nextRemaining, completed := patchRemainingIDs(payload.RemainingIDs, completedAchievementID)

		nextStatusCode := models.AchievementStatusInProgress
		if completed {
			nextStatusCode = models.AchievementStatusIssued
		}

		nextStatus, err := s.repo.GetAchievementStatusByCode(ctx, nextStatusCode)
		if err != nil {
			return fmt.Errorf("get next status: %w", err)
		}
		if nextStatus == nil {
			return services.ErrStatusNotFound
		}

		nextPayload, err := json.Marshal(InProgressPayload{
			RemainingIDs: nextRemaining,
		})
		if err != nil {
			return err
		}

		if err := s.repo.UpdateIssuanceProgress(ctx, row.ID, nextStatus.ID, nextPayload); err != nil {
			return fmt.Errorf("update issuance progress: %w", err)
		}
	}

	return nil
}

func (s *Service) bootstrapDependents(
	ctx context.Context,
	issuedAchievement *models.Achievement,
	recipientID uuid.UUID,
	issuerID uuid.UUID,
	inProgressStatusID int64,
) error {
	dependents, err := s.repo.FindDependentAchievements(ctx, issuedAchievement.ID, issuedAchievement.CohortID)
	if err != nil {
		return fmt.Errorf("find dependent achievements: %w", err)
	}

	for _, dependent := range dependents {
		// Safety: skip if recipient already has this parent achievement.
		existing, err := s.repo.GetRecipientIssuance(ctx, dependent.ID, recipientID)
		if err != nil {
			return fmt.Errorf("get dependent recipient issuance: %w", err)
		}
		if existing != nil {
			continue
		}

		conditionType, err := s.repo.GetConditionTypeByID(ctx, dependent.ConditionTypeID)
		if err != nil {
			return fmt.Errorf("get condition type: %w", err)
		}
		if conditionType == nil {
			return services.ErrConditionTypeNotFound
		}

		allOfPayload, err := parseConditionPayload(conditionType.Code, dependent.ConditionPayload)
		if err != nil {
			return err
		}

		nextRemaining, completed := patchRemainingIDs(allOfPayload.AchievementIDs, issuedAchievement.ID)

		if completed {
			if _, err := s.createDirectIssuance(ctx, dependent.ID, recipientID, issuerID, nil); err != nil {
				return err
			}
			continue
		}

		progressPayload, err := json.Marshal(InProgressPayload{
			RemainingIDs: nextRemaining,
		})
		if err != nil {
			return err
		}

		_, err = s.repo.IssueAchievement(ctx, models.AchievementIssuance{
			AchievementID:   dependent.ID,
			RecipientID:     recipientID,
			IssuerID:        issuerID,
			StatusID:        inProgressStatusID,
			ProgressPayload: progressPayload,
		})
		if err != nil {
			return fmt.Errorf("create in-progress dependent issuance: %w", err)
		}
	}

	return nil
}

func parseConditionPayload(conditionTypeCode string, raw json.RawMessage) (*AllOfConditionPayload, error) {
	switch conditionTypeCode {
	case models.ConditionTypeAllOf:
		return parseAllOfPayload(raw)
	default:
		return nil, services.ErrConditionTypeNotFound
	}
}

func parseAllOfPayload(raw json.RawMessage) (*AllOfConditionPayload, error) {
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
	return &payload, nil
}

func patchRemainingIDs(source []int64, completedID int64) ([]int64, bool) {
	next := make([]int64, 0, len(source))
	for _, id := range source {
		if id != completedID {
			next = append(next, id)
		}
	}
	return next, len(next) == 0
}
