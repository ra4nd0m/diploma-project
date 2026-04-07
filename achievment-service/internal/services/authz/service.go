package authz

import (
	"achievement-service/internal/services"
	"context"
	"errors"

	"github.com/google/uuid"
)

type cohortAccessChecker interface {
	CanEditCohort(ctx context.Context, userID uuid.UUID, cohortID int64) (bool, error)
	IsUserInCohort(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]int64, error)
}

type Service struct {
	cohortChecker cohortAccessChecker
}

func NewService(cohortChecker cohortAccessChecker) *Service {
	return &Service{
		cohortChecker: cohortChecker,
	}
}

func (s *Service) RequireCohortEditAccess(
	ctx context.Context,
	userID uuid.UUID,
	cohortID int64,
) error {
	if userID == uuid.Nil || cohortID <= 0 {
		return errors.New("invalid authorization input")
	}

	allowed, err := s.cohortChecker.CanEditCohort(ctx, userID, cohortID)
	if err != nil {
		return err
	}
	if !allowed {
		return services.ErrForbidden
	}

	return nil
}

func (s *Service) RequireUserInCohorts(
	ctx context.Context,
	userID uuid.UUID,
	cohortIDs []int64,
) ([]int64, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid authorization input")
	}

	uniqueRequested := uniquePositive(cohortIDs)
	if len(uniqueRequested) == 0 {
		return []int64{}, nil
	}

	allowedIDs, err := s.cohortChecker.IsUserInCohort(ctx, userID, uniqueRequested)
	if err != nil {
		return nil, err
	}

	allowedSet := make(map[int64]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		if id <= 0 {
			continue
		}
		allowedSet[id] = struct{}{}
	}

	for _, requestedID := range uniqueRequested {
		if _, ok := allowedSet[requestedID]; !ok {
			return nil, services.ErrForbidden
		}
	}

	return uniqueRequested, nil
}

func uniquePositive(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}
