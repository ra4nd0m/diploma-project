package achievementissue

import "github.com/google/uuid"

type Input struct {
	AchievementID    int64
	RecipientID      uuid.UUID
	IssuerID         uuid.UUID
	AdditionalDetail *string
}

type InProgressPayload struct {
	RemainingIDs []int64 `json:"remaining_ids"`
}

type AllOfConditionPayload struct {
	AchievementIDs []int64 `json:"achievement_ids"`
}

type Output struct {
	ID int64
}
