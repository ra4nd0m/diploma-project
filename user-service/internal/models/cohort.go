package models

import "github.com/google/uuid"

type Cohort struct {
	ID      int64
	Name    string
	OwnerID uuid.UUID
}

type CohortWithUsers struct {
	ID      int64
	Name    string
	OwnerID uuid.UUID
	Users   []User
}
