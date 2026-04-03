package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type AchievementRepo struct {
	db *sql.DB
}

func NewAchievementRepo(db *sql.DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

func (r *AchievementRepo) GetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			icon_link,
			owner_id,
			cohort_id,
			access_mode,
			issuance_kind,
			condition_type,
			condition_payload
		FROM achievement
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, achievementID)
	achievement, err := scanAchievement(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get achievement by id %d: %w", achievementID, err)
	}

	return achievement, nil
}

func (r *AchievementRepo) GetAchievements(ctx context.Context, userID uuid.UUID) ([]*models.Achievement, error) {
	const query = `
		SELECT
			id,
			name,
			description,
			icon_link,
			owner_id,
			cohort_id,
			access_mode,
			issuance_kind,
			condition_type,
			condition_payload
		FROM achievement
		WHERE owner_id = $1
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get achievements by owner %s: %w", userID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	achievements := make([]*models.Achievement, 0)
	for rows.Next() {
		achievement, err := scanAchievement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan achievements by owner %s: %w", userID, err)
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements by owner %s: %w", userID, err)
	}

	return achievements, nil
}

func (r *AchievementRepo) CreateAchievement(ctx context.Context, achievement models.Achievement) (int64, error) {
	const query = `
		INSERT INTO achievement (
			name,
			description,
			icon_link,
			owner_id,
			cohort_id,
			access_mode,
			issuance_kind,
			condition_type,
			condition_payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var conditionType any
	if achievement.ConditionTypeID > 0 {
		conditionType = achievement.ConditionTypeID
	}

	var conditionPayload any
	if len(achievement.ConditionPayload) > 0 {
		conditionPayload = achievement.ConditionPayload
	}

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		achievement.Name,
		achievement.Description,
		achievement.IconLink,
		achievement.OwnerID,
		achievement.CohortID,
		achievement.AccessModeID,
		achievement.IssuanceKindID,
		conditionType,
		conditionPayload,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create achievement: %w", err)
	}

	return id, nil
}

func scanAchievement(scanner interface{ Scan(dest ...any) error }) (*models.Achievement, error) {
	var achievement models.Achievement
	var conditionType sql.NullInt64
	var conditionPayload []byte

	err := scanner.Scan(
		&achievement.ID,
		&achievement.Name,
		&achievement.Description,
		&achievement.IconLink,
		&achievement.OwnerID,
		&achievement.CohortID,
		&achievement.AccessModeID,
		&achievement.IssuanceKindID,
		&conditionType,
		&conditionPayload,
	)
	if err != nil {
		return nil, fmt.Errorf("scan achievement: %w", err)
	}

	if conditionType.Valid {
		achievement.ConditionTypeID = conditionType.Int64
	}
	if len(conditionPayload) > 0 {
		achievement.ConditionPayload = conditionPayload
	}

	return &achievement, nil
}

func (r *AchievementRepo) FindDependentAchievements(ctx context.Context, dependencyAchievementID, cohortID int64) ([]*models.Achievement, error) {
	const query = `
		SELECT DISTINCT
			a.id,
			a.name,
			a.description,
			a.icon_link,
			a.owner_id,
			a.cohort_id,
			a.access_mode,
			a.issuance_kind,
			a.condition_type,
			a.condition_payload
		FROM achievement a
		JOIN condition_type ct ON ct.id = a.condition_type
		CROSS JOIN jsonb_array_elements(a.condition_payload->'achievement_ids') AS dep(id_str)
		WHERE a.cohort_id = $1
			AND ct.code = $2
			AND a.condition_payload IS NOT NULL
			AND dep.id_str::bigint = $3
	`
	rows, err := r.db.QueryContext(ctx, query, cohortID, models.ConditionTypeAllOf, dependencyAchievementID)
	if err != nil {
		return nil, fmt.Errorf("find dependent achievements for dependency %d in cohort %d: %w", dependencyAchievementID, cohortID, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	achievements := make([]*models.Achievement, 0)
	for rows.Next() {
		achievement, err := scanAchievement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan dependent achievements for dependency %d in cohort %d: %w", dependencyAchievementID, cohortID, err)
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependent achievements for dependency %d in cohort %d: %w", dependencyAchievementID, cohortID, err)
	}

	return achievements, nil
}
