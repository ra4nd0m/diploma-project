package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type IssuanceLifecycleRepo struct {
	db *sql.DB
}

func NewIssuanceLifecycleRepo(db *sql.DB) *IssuanceLifecycleRepo {
	return &IssuanceLifecycleRepo{db: db}
}

func (r *IssuanceLifecycleRepo) IssueAchievement(ctx context.Context, issuance models.AchievementIssuance) (int64, error) {
	const query = `
		INSERT INTO achievement_issuance (
			achievement_id,
			recipient_id,
			issuer_id,
			"status",
			additional_detail,
			progress_payload
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(
		ctx,
		query,
		issuance.AchievementID,
		issuance.RecipientID,
		issuance.IssuerID,
		issuance.StatusID,
		issuance.AdditionalDetail,
		issuance.ProgressPayload,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("issue achievement: %w", err)
	}

	return id, nil
}

func (r *IssuanceLifecycleRepo) GetRecipientIssuance(ctx context.Context, achievementID int64, recipientID uuid.UUID) (*models.AchievementIssuance, error) {
	const query = `
		SELECT
			id,
			achievement_id,
			recipient_id,
			issuer_id,
			"status",
			additional_detail,
			progress_payload
		FROM achievement_issuance
		WHERE achievement_id = $1
			AND recipient_id = $2
	`

	var issuance models.AchievementIssuance
	var additionalDetail sql.NullString
	var progressPayload []byte

	err := r.db.QueryRowContext(ctx, query, achievementID, recipientID).Scan(
		&issuance.ID,
		&issuance.AchievementID,
		&issuance.RecipientID,
		&issuance.IssuerID,
		&issuance.StatusID,
		&additionalDetail,
		&progressPayload,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get recipient issuance for achievement %d and recipient %s: %w", achievementID, recipientID, err)
	}

	if additionalDetail.Valid {
		issuance.AdditionalDetail = &additionalDetail.String
	}
	if len(progressPayload) > 0 {
		issuance.ProgressPayload = progressPayload
	}

	return &issuance, nil
}

func (r *IssuanceLifecycleRepo) FindDependentsByStatus(ctx context.Context, recipientID uuid.UUID, cohortID, statusID, dependencyAchievementID int64) ([]*models.AchievementIssuance, error) {
	const query = `
		SELECT
			ai.id,
			ai.achievement_id,
			ai.recipient_id,
			ai.issuer_id,
			ai."status",
			ai.additional_detail,
			ai.progress_payload
		FROM achievement_issuance ai
		JOIN achievement a ON ai.achievement_id = a.id
		WHERE ai.recipient_id = $1
			AND a.cohort_id = $2
			AND ai."status" = $3
			and ai.progress_payload is not null
			and ai.progress_payload @> jsonb_build_object(
				'remaining_ids',
				jsonb_build_array($4)
			)
	`
	rows, err := r.db.QueryContext(ctx, query, recipientID, cohortID, statusID, dependencyAchievementID)
	if err != nil {
		return nil, fmt.Errorf("find dependents by status: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()

	var issuances []*models.AchievementIssuance
	for rows.Next() {
		var issuance models.AchievementIssuance
		var additionalDetail sql.NullString
		var progressPayload []byte

		if err := rows.Scan(
			&issuance.ID,
			&issuance.AchievementID,
			&issuance.RecipientID,
			&issuance.IssuerID,
			&issuance.StatusID,
			&additionalDetail,
			&progressPayload,
		); err != nil {
			return nil, fmt.Errorf("scan dependent issuance: %w", err)
		}

		if additionalDetail.Valid {
			issuance.AdditionalDetail = &additionalDetail.String
		}
		if len(progressPayload) > 0 {
			issuance.ProgressPayload = progressPayload
		}

		issuances = append(issuances, &issuance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependent issuances: %w", err)
	}

	return issuances, nil
}

func (r *IssuanceLifecycleRepo) FindDependentAchievements(ctx context.Context, dependencyAchievementID, cohortID int64, conditionTypeCode string) ([]*models.Achievement, error) {
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

	rows, err := r.db.QueryContext(ctx, query, cohortID, conditionTypeCode, dependencyAchievementID)
	if err != nil {
		return nil, fmt.Errorf("find dependent achievements for dependency %d in cohort %d and condition type %s: %w", dependencyAchievementID, cohortID, conditionTypeCode, err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}()

	var achievements []*models.Achievement
	for rows.Next() {
		var achievement models.Achievement
		var conditionType sql.NullInt64
		var conditionPayload []byte

		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("scan dependent achievements: %w", err)
		}

		if conditionType.Valid {
			achievement.ConditionTypeID = conditionType.Int64
		}
		if len(conditionPayload) > 0 {
			achievement.ConditionPayload = conditionPayload
		}

		achievements = append(achievements, &achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependent achievements: %w", err)
	}

	return achievements, nil
}

func (r *IssuanceLifecycleRepo) UpdateIssuanceProgress(ctx context.Context, issuanceID int64, statusID int64, progressPayload []byte) error {
	const query = `
		UPDATE achievement_issuance
		SET "status" = $1,
			progress_payload = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, statusID, progressPayload, issuanceID)
	if err != nil {
		return fmt.Errorf("update issuance progress for issuance %d: %w", issuanceID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected for issuance %d progress update: %w", issuanceID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issuance %d not found", issuanceID)
	}

	return nil
}
