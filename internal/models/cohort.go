package models

import "github.com/google/uuid"

type Cohort struct {
	ID      uuid.UUID
	Name    string
	OwnerID uuid.UUID
}
