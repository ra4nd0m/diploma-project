package models

type AccessMode struct {
	ID   int64
	Code string
	Name string
}

type ConditionType struct {
	ID   int64
	Code string
	Name string
}

type IssuanceKind struct {
	ID   int64
	Code string
	Name string
}

type AchievementStatus struct {
	ID   int64
	Code string
	Name string
}
