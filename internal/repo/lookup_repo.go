package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"
)

type LookupRepo struct {
	db *sql.DB
}

func NewLookupRepo(db *sql.DB) *LookupRepo {
	return &LookupRepo{db: db}
}

func (r *LookupRepo) GetAccessModeByCode(ctx context.Context, code string) (*models.AccessMode, error) {
	const query = `SELECT id, code, "name" FROM access_mode WHERE code = $1`
	var accessMode models.AccessMode
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&accessMode.ID, &accessMode.Code, &accessMode.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &accessMode, nil
}

func (r *LookupRepo) GetAccessModeByID(ctx context.Context, id int64) (*models.AccessMode, error) {
	const query = `SELECT id, code, "name" FROM access_mode WHERE id = $1`
	var accessMode models.AccessMode
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&accessMode.ID, &accessMode.Code, &accessMode.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &accessMode, nil
}

func (r *LookupRepo) GetIssuanceKindByCode(ctx context.Context, code string) (*models.IssuanceKind, error) {
	const query = `SELECT id, code, "name" FROM issuance_kind WHERE code = $1`
	var issuanceKind models.IssuanceKind
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&issuanceKind.ID, &issuanceKind.Code, &issuanceKind.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &issuanceKind, nil
}

func (r *LookupRepo) GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error) {
	const query = `SELECT id, code, "name" FROM issuance_kind WHERE id = $1`
	var issuanceKind models.IssuanceKind
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&issuanceKind.ID, &issuanceKind.Code, &issuanceKind.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &issuanceKind, nil
}

func (r *LookupRepo) GetConditionTypeByCode(ctx context.Context, code string) (*models.ConditionType, error) {
	const query = `SELECT id, code, "name" FROM condition_type WHERE code = $1`
	var conditionType models.ConditionType
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&conditionType.ID, &conditionType.Code, &conditionType.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &conditionType, nil
}

func (r *LookupRepo) GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error) {
	const query = `SELECT id, code, "name" FROM condition_type WHERE id = $1`
	var conditionType models.ConditionType
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&conditionType.ID, &conditionType.Code, &conditionType.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &conditionType, nil
}

func (r *LookupRepo) GetAchievementStatusByCode(ctx context.Context, code string) (*models.AchievementStatus, error) {
	const query = `SELECT id, code, "name" FROM achievement_status WHERE code = $1`
	var status models.AchievementStatus
	if err := r.db.QueryRowContext(ctx, query, code).Scan(&status.ID, &status.Code, &status.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &status, nil
}
