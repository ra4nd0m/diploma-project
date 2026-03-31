package repo

import (
	"context"
	"database/sql"
	"fmt"
	"user-service/internal/models"

	"github.com/google/uuid"
)

type CohortRepo struct {
	db *sql.DB
}

func NewCohortRepo(db *sql.DB) *CohortRepo {
	return &CohortRepo{db: db}
}

func (r *CohortRepo) CreateCohort(ctx context.Context, name string, ownerID uuid.UUID) (*models.Cohort, error) {
	const query = `
		INSERT INTO cohort (name, owner_id)
		VALUES ($1, $2)
		RETURNING id, name, owner_id
	`

	var cohort models.Cohort

	err := r.db.QueryRowContext(ctx, query, name, ownerID).Scan(
		&cohort.ID,
		&cohort.Name,
		&cohort.OwnerID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert cohort %w", err)
	}

	return &cohort, nil
}

func (r *CohortRepo) GetCohortByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*models.Cohort, error) {
	const query = `
		SELECT id, name, owner_id
		FROM cohort
		WHERE owner_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query cohorts by owner id %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	cohorts := make([]*models.Cohort, 0)

	for rows.Next() {
		var cohort models.Cohort

		err := rows.Scan(
			&cohort.ID,
			&cohort.Name,
			&cohort.OwnerID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cohort %w", err)
		}

		cohorts = append(cohorts, &cohort)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cohort rows %w", err)
	}

	return cohorts, nil
}

func (r *CohortRepo) GetCohortList(ctx context.Context) ([]*models.Cohort, error) {
	const query = `
		SELECT id, name, owner_id
		FROM cohort
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query cohorts %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	cohorts := make([]*models.Cohort, 0)

	for rows.Next() {
		var cohort models.Cohort

		err := rows.Scan(
			&cohort.ID,
			&cohort.Name,
			&cohort.OwnerID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cohort %w", err)
		}

		cohorts = append(cohorts, &cohort)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cohort rows %w", err)
	}

	return cohorts, nil

}

func (r *CohortRepo) GetCohortByID(ctx context.Context, id int64) (*models.CohortWithUsers, bool, error) {
	const query = `
		SELECT c.id, c.name, c.owner_id, u.id, u.display_name
		FROM cohort c
		LEFT JOIN user_cohort uc ON c.id = uc.cohort_id
		LEFT JOIN "user" u ON uc.user_id = u.id
		WHERE c.id = $1
		ORDER BY u.display_name NULLS LAST, u.id
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, false, fmt.Errorf("query cohort by id %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("Error closing rows: %v\n", err)
		}
	}()

	var result *models.CohortWithUsers

	for rows.Next() {
		var (
			cohortID    int64
			name        string
			ownerID     uuid.UUID
			userID      *uuid.UUID
			displayName *string
		)

		err := rows.Scan(
			&cohortID,
			&name,
			&ownerID,
			&userID,
			&displayName,
		)
		if err != nil {
			return nil, false, fmt.Errorf("scan cohort with users %w", err)
		}

		if result == nil {
			result = &models.CohortWithUsers{
				ID:      cohortID,
				Name:    name,
				OwnerID: ownerID,
				Users:   make([]models.User, 0),
			}
		}

		if userID != nil {
			member := models.User{
				ID: *userID,
			}
			if displayName != nil {
				member.DisplayName = *displayName
			}

			result.Users = append(result.Users, member)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate cohort with users rows %w", err)
	}

	if result == nil {
		return nil, false, nil
	}

	return result, true, nil
}

func (r *CohortRepo) AddUserToCohort(ctx context.Context, cohortID int64, userID uuid.UUID) error {
	const query = `
		INSERT INTO user_cohort (cohort_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (cohort_id, user_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, cohortID, userID)
	if err != nil {
		return fmt.Errorf("insert user cohort %w", err)
	}

	return nil
}

func (r *CohortRepo) RemoveUserFromCohort(ctx context.Context, cohortID int64, userID uuid.UUID) error {
	const query = `
		DELETE FROM user_cohort
		WHERE cohort_id = $1 AND user_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, cohortID, userID)
	if err != nil {
		return fmt.Errorf("delete user cohort %w", err)
	}

	return nil
}
