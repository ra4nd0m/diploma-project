package services

import "errors"

var (
	ErrForbidden             = errors.New("forbidden")
	ErrInvalidInput          = errors.New("invalid input")
	ErrAccessModeNotFound    = errors.New("access mode not found")
	ErrIssuanceKindNotFound  = errors.New("issuance kind not found")
	ErrConditionTypeNotFound = errors.New("condition type not found")
	ErrInvalidCondition      = errors.New("invalid condition")
	ErrNotFound              = errors.New("not found")
)
