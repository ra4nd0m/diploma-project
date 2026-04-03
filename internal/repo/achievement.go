package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"

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
		return nil, err
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
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	achievements := make([]*models.Achievement, 0)
	for rows.Next() {
		achievement, err := scanAchievement(rows)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
		return 0, err
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
		return nil, err
	}

	if conditionType.Valid {
		achievement.ConditionTypeID = conditionType.Int64
	}
	if len(conditionPayload) > 0 {
		achievement.ConditionPayload = conditionPayload
	}

	return &achievement, nil

}
