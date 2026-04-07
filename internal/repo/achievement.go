package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type AchievementRepo struct {
	db *sql.DB
}

func NewAchievementRepo(db *sql.DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

func (r *AchievementRepo) GetAchievementsByOwner(
	ctx context.Context,
	userID uuid.UUID,
	cohortIDs []int64,
) ([]*models.Achievement, error) {
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
			AND (
					cardinality($2::bigint[]) = 0
					OR cohort_id = ANY($2::bigint[])
			)
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, cohortIDs)
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

func (r *AchievementRepo) ListVisibleAchievements(
	ctx context.Context,
	userID uuid.UUID,
	cohortIDs []int64,
	publicAccessModeID int64,
	cohortAccessModeID int64,
	privateAccessModeID int64,
) ([]*models.Achievement, error) {
	const query = `
		SELECT
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
		WHERE 
			a.access_mode = $1
			OR (
				a.access_mode = $2
				AND a.cohort_id = ANY($3::bigint[])
			) 
			OR (
				a.access_mode = $4
				AND a.owner_id = $5
			)
		ORDER BY a.id DESC
	`
	rows, err := r.db.QueryContext(
		ctx,
		query,
		publicAccessModeID,
		cohortAccessModeID,
		cohortIDs,
		privateAccessModeID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list visible achievements for user %s in cohorts %v: %w", userID, cohortIDs, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	achievements := make([]*models.Achievement, 0)
	for rows.Next() {
		achievement, err := scanAchievement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan visible achievements for user %s in cohorts %v: %w", userID, cohortIDs, err)
		}
		achievements = append(achievements, achievement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate visible achievements for user %s in cohorts %v: %w", userID, cohortIDs, err)
	}

	return achievements, nil
}

func (r *AchievementRepo) ListAchievementsForRecipient(
	ctx context.Context,
	requestUserId uuid.UUID,
	recipientID uuid.UUID,
	cohortIDs []int64,
	publicAccessModeID int64,
	cohortAccessModeID int64,
	privateAccessModeID int64,
) ([]*models.PersonalAchievement, error) {
	const query = `
		SELECT
			a.id,
			a.name,
			a.description,
			a.icon_link,
			a.owner_id,
			a.cohort_id,
			a.access_mode,
			a.issuance_kind,
			a.condition_type,
			a.condition_payload,

			ai.id,
			ai.status,
			s.code,
			ai.additional_detail,
			ai.progress_payload,
		FROM achievement a
		LEFT JOIN achievement_issue ai 
			ON ai.achievement_id = a.id 
			AND ai.recipient_id = $1
		LEFT JOIN achievement_status s
			ON s.id = ai.status
		WHERE
			a.access_mode = $2
			OR (
				a.access_mode = $3
				AND a.cohort_id = ANY($4::bigint[])
			)
			OR (
				a.access_mode = $5
				AND a.owner_id = $6
			)
		ORDER BY a.id DESC, ai.id DESC
	`
	rows, err := r.db.QueryContext(
		ctx,
		query,
		recipientID,
		publicAccessModeID,
		cohortAccessModeID,
		cohortIDs,
		privateAccessModeID,
		requestUserId,
	)

	if err != nil {
		return nil, fmt.Errorf("list achievements for recipient %s requested by user %s in cohorts %v: %w", recipientID, requestUserId, cohortIDs, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	achievements := make([]*models.PersonalAchievement, 0)
	for rows.Next() {
		item, err := scanPersonalAchievement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan achievements for recipient %s requested by user %s in cohorts %v: %w", recipientID, requestUserId, cohortIDs, err)
		}
		achievements = append(achievements, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements for recipient %s requested by user %s in cohorts %v: %w", recipientID, requestUserId, cohortIDs, err)
	}

	return achievements, nil
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

func scanPersonalAchievement(scanner interface{ Scan(dest ...any) error }) (*models.PersonalAchievement, error) {
	var item models.PersonalAchievement

	var conditionType sql.NullInt64
	var conditionPayload []byte

	var issuanceID sql.NullInt64
	var statusID sql.NullInt64
	var statusCode sql.NullString
	var progressPayload []byte
	var issuanceCreatedOn sql.NullTime

	err := scanner.Scan(
		&item.AchievementID,
		&item.Name,
		&item.Description,
		&item.IconLink,
		&item.OwnerID,
		&item.CohortID,
		&item.AccessModeID,
		&item.IssuanceKindID,
		&conditionType,
		&conditionPayload,

		&issuanceID,
		&statusID,
		&statusCode,
		&item.AdditionalDetail,
		&progressPayload,
		&issuanceCreatedOn,
	)
	if err != nil {
		return nil, fmt.Errorf("scan personal achievement: %w", err)
	}

	if conditionType.Valid {
		item.ConditionTypeID = conditionType.Int64
	}
	if len(conditionPayload) > 0 {
		item.ConditionPayload = json.RawMessage(conditionPayload)
	}

	if issuanceID.Valid {
		v := issuanceID.Int64
		item.IssuanceID = &v
	}
	if statusID.Valid {
		v := statusID.Int64
		item.StatusID = &v
	}
	if statusCode.Valid {
		v := statusCode.String
		item.StatusCode = &v
	}
	if len(progressPayload) > 0 {
		item.ProgressPayload = json.RawMessage(progressPayload)
	}

	return &item, nil
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
