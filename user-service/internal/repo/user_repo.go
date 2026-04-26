package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"user-service/internal/models"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, bool, error) {
	const query = `
		SELECT id, display_name, preferences
		FROM "user"
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.DisplayName,
		&user.Preferences,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("query user by id %w", err)
	}

	return &user, true, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	const query = `
		INSERT INTO "user" (id, display_name, preferences)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.DisplayName, user.Preferences)
	if err != nil {
		return fmt.Errorf("insert user %w", err)
	}

	return nil
}

func (r *UserRepo) UpdateUserPreferences(ctx context.Context, id uuid.UUID, preferences json.RawMessage) error {
	const query = `
		UPDATE "user"
		SET preferences = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, preferences)
	if err != nil {
		return fmt.Errorf("update user preferences %w", err)
	}

	return nil
}
