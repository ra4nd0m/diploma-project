package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("forbidden")

type cohortAccessChecker interface {
	CanEditCohort(ctx context.Context, userID uuid.UUID, cohortID int64) (bool, error)
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
		return ErrForbidden
	}

	return nil
}
