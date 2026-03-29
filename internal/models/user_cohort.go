package models

import (
	"time"

	"github.com/google/uuid"
)

type UserCohort struct {
	UserID   uuid.UUID `db:"user_id"`
	CohortID uuid.UUID `db:"cohort_id"`
	JoinedOn time.Time `db:"joined_on"`
}
