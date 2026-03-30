package services

import (
	"context"
	"encoding/json"
	"fmt"
	"user-service/internal/models"

	"github.com/google/uuid"
)

type UserRepo interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, bool, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUserPreferences(ctx context.Context, id uuid.UUID, preferences json.RawMessage) error
}

type UserService struct {
	users UserRepo
}

func NewUserService(users UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetOrCreateUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, found, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id %w", err)
	}
	if found {
		return user, nil
	} else {
		newUser := &models.User{
			ID:          id,
			DisplayName: id.String(),
			Preferences: json.RawMessage(`{}`),
		}
		err = s.users.CreateUser(ctx, newUser)
		if err != nil {
			return nil, fmt.Errorf("create user %w", err)
		}
		return newUser, nil
	}
}
