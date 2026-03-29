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
