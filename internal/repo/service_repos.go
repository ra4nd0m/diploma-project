package repo

import (
	"achievement-service/internal/models"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// AchievementCreationServiceRepo composes split repositories for creation service needs.
type AchievementCreationServiceRepo struct {
	achievementRepo *AchievementRepo
	lookupRepo      *LookupRepo
}

func NewAchievementCreationServiceRepo(db *sql.DB) *AchievementCreationServiceRepo {
	return &AchievementCreationServiceRepo{
		achievementRepo: NewAchievementRepo(db),
		lookupRepo:      NewLookupRepo(db),
	}
}

func (r *AchievementCreationServiceRepo) GetAccessModeByCode(ctx context.Context, code string) (*models.AccessMode, error) {
	return r.lookupRepo.GetAccessModeByCode(ctx, code)
}

func (r *AchievementCreationServiceRepo) GetIssuanceKindByCode(ctx context.Context, code string) (*models.IssuanceKind, error) {
	return r.lookupRepo.GetIssuanceKindByCode(ctx, code)
}

func (r *AchievementCreationServiceRepo) GetConditionTypeByCode(ctx context.Context, code string) (*models.ConditionType, error) {
	return r.lookupRepo.GetConditionTypeByCode(ctx, code)
}

func (r *AchievementCreationServiceRepo) CreateAchievement(ctx context.Context, achievement models.Achievement) (int64, error) {
	return r.achievementRepo.CreateAchievement(ctx, achievement)
}

// AchievementReadingServiceRepo composes split repositories for reading service needs.
type AchievementReadingServiceRepo struct {
	achievementRepo *AchievementRepo
	lookupRepo      *LookupRepo
}

func NewAchievementReadingServiceRepo(db *sql.DB) *AchievementReadingServiceRepo {
	return &AchievementReadingServiceRepo{
		achievementRepo: NewAchievementRepo(db),
		lookupRepo:      NewLookupRepo(db),
	}
}

func (r *AchievementReadingServiceRepo) GetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error) {
	return r.achievementRepo.GetAchievement(ctx, achievementID)
}

func (r *AchievementReadingServiceRepo) GetAchievements(ctx context.Context, userID uuid.UUID) ([]*models.Achievement, error) {
	return r.achievementRepo.GetAchievements(ctx, userID)
}

func (r *AchievementReadingServiceRepo) GetAccessModeByID(ctx context.Context, id int64) (*models.AccessMode, error) {
	return r.lookupRepo.GetAccessModeByID(ctx, id)
}

func (r *AchievementReadingServiceRepo) GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error) {
	return r.lookupRepo.GetIssuanceKindByID(ctx, id)
}

func (r *AchievementReadingServiceRepo) GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error) {
	return r.lookupRepo.GetConditionTypeByID(ctx, id)
}

// AchievementIssueServiceRepo composes split repositories for issuance service needs.
type AchievementIssueServiceRepo struct {
	achievementRepo       *AchievementRepo
	lookupRepo            *LookupRepo
	issuanceLifecycleRepo *IssuanceLifecycleRepo
}

func NewAchievementIssueServiceRepo(db *sql.DB) *AchievementIssueServiceRepo {
	return &AchievementIssueServiceRepo{
		achievementRepo:       NewAchievementRepo(db),
		lookupRepo:            NewLookupRepo(db),
		issuanceLifecycleRepo: NewIssuanceLifecycleRepo(db),
	}
}

func (r *AchievementIssueServiceRepo) GetAchievement(ctx context.Context, achievementID int64) (*models.Achievement, error) {
	return r.achievementRepo.GetAchievement(ctx, achievementID)
}

func (r *AchievementIssueServiceRepo) GetIssuanceKindByID(ctx context.Context, id int64) (*models.IssuanceKind, error) {
	return r.lookupRepo.GetIssuanceKindByID(ctx, id)
}

func (r *AchievementIssueServiceRepo) GetConditionTypeByID(ctx context.Context, id int64) (*models.ConditionType, error) {
	return r.lookupRepo.GetConditionTypeByID(ctx, id)
}

func (r *AchievementIssueServiceRepo) GetAchievementStatusByCode(ctx context.Context, code string) (*models.AchievementStatus, error) {
	return r.lookupRepo.GetAchievementStatusByCode(ctx, code)
}

func (r *AchievementIssueServiceRepo) GetRecipientIssuance(ctx context.Context, achievementID int64, recipientID uuid.UUID) (*models.AchievementIssuance, error) {
	return r.issuanceLifecycleRepo.GetRecipientIssuance(ctx, achievementID, recipientID)
}

func (r *AchievementIssueServiceRepo) IssueAchievement(ctx context.Context, issuance models.AchievementIssuance) (int64, error) {
	return r.issuanceLifecycleRepo.IssueAchievement(ctx, issuance)
}

func (r *AchievementIssueServiceRepo) FindDependentsByStatus(ctx context.Context, recipientID uuid.UUID, cohortID, statusID, dependencyAchievementID int64) ([]*models.AchievementIssuance, error) {
	return r.issuanceLifecycleRepo.FindDependentsByStatus(ctx, recipientID, cohortID, statusID, dependencyAchievementID)
}

func (r *AchievementIssueServiceRepo) FindDependentAchievements(ctx context.Context, dependencyAchievementID int64, cohortID int64, conditionTypeCode string) ([]*models.Achievement, error) {
	return r.issuanceLifecycleRepo.FindDependentAchievements(ctx, dependencyAchievementID, cohortID, conditionTypeCode)
}

func (r *AchievementIssueServiceRepo) UpdateIssuanceProgress(ctx context.Context, issuanceID int64, statusID int64, progressPayload []byte) error {
	return r.issuanceLifecycleRepo.UpdateIssuanceProgress(ctx, issuanceID, statusID, progressPayload)
}
