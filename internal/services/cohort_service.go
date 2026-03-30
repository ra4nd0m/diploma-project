package services

import (
	"context"
	"fmt"
	"user-service/internal/models"

	"github.com/google/uuid"
)

type CohortRepo interface {
	CreateCohort(ctx context.Context, name string, ownerID uuid.UUID) (*models.Cohort, error)
	GetCohortByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*models.Cohort, error)
	GetCohortByID(ctx context.Context, id uuid.UUID) (*models.CohortWithUsers, bool, error)
	AddUserToCohort(ctx context.Context, cohortID, userID uuid.UUID) error
	GetCohortList(ctx context.Context) ([]*models.Cohort, error)
}

type InviteTokenManager interface {
	GenerateInviteToken(cohortID string) (string, error)
}

type UserProvider interface {
	GetOrCreateUser(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type CohortService struct {
	cohorts      CohortRepo
	tokenManager InviteTokenManager
	users        UserProvider
}

func NewCohortService(cohorts CohortRepo, tokenManager InviteTokenManager, users UserProvider) *CohortService {
	return &CohortService{cohorts: cohorts, tokenManager: tokenManager, users: users}
}

func (s *CohortService) CreateCohort(ctx context.Context, name string, ownerID uuid.UUID) (*models.Cohort, error) {
	cohort, err := s.cohorts.CreateCohort(ctx, name, ownerID)
	if err != nil {
		return nil, fmt.Errorf("create cohort %w", err)
	}
	return cohort, nil
}

func (s *CohortService) GetCohorts(ctx context.Context) ([]*models.Cohort, error) {
	cohorts, err := s.cohorts.GetCohortList(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cohorts %w", err)
	}
	return cohorts, nil
}

func (s *CohortService) GetCohortWithUsers(ctx context.Context, cohortID uuid.UUID) (*models.CohortWithUsers, error) {
	cohort, found, err := s.cohorts.GetCohortByID(ctx, cohortID)
	if err != nil {
		return nil, fmt.Errorf("get cohort with users %w", err)
	}
	if !found {
		return nil, fmt.Errorf("cohort not found")
	}
	return cohort, nil
}

func (s *CohortService) AddsUserToCohortByInvite(ctx context.Context, cohortID, userID uuid.UUID) error {
	user, err := s.users.GetOrCreateUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get or create user %w", err)
	}
	err = s.cohorts.AddUserToCohort(ctx, cohortID, user.ID)
	if err != nil {
		return fmt.Errorf("add user to cohort %w", err)
	}
	return nil
}

func (s *CohortService) GenerateInviteTokenToCohort(ctx context.Context, cohortID uuid.UUID) (string, error) {
	token, err := s.tokenManager.GenerateInviteToken(cohortID.String())
	if err != nil {
		return "", fmt.Errorf("generate invite token %w", err)
	}
	return token, nil
}
