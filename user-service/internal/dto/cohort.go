package dto

// CohortResponse represents basic cohort information.
type CohortResponse struct {
	ID   string `json:"id" example:"123"`
	Name string `json:"name" example:"Advanced Mathematics"`
}

// CohortWithUsersResponse represents a cohort with all its member users.
type CohortWithUsersResponse struct {
	ID    string         `json:"id" example:"123"`
	Name  string         `json:"name" example:"Advanced Mathematics"`
	Users []UserResponse `json:"users"`
}

// CohortCreateRequest represents a request to create a new cohort.
type CohortCreateRequest struct {
	Name string `json:"name" binding:"required" example:"Advanced Mathematics"`
}

// CohortJoinRequest represents a request to join a cohort using an invite token.
type CohortJoinRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// CohortIsOwnedRequest represents a request to check if a user owns a cohort.
type CohortIsOwnedRequest struct {
	CohortID string `json:"cohort_id" binding:"required" example:"123"`
	UserID   string `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CohortIsUserInRequest represents a request to check which cohorts a user is a member of.
type CohortIsUserInRequest struct {
	UserID    string  `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CohortIDs []int64 `json:"cohort_ids" binding:"required" example:"1,2,3"`
}

// CohortIsUserInResponse represents the response with cohort IDs a user is a member of.
type CohortIsUserInResponse struct {
	CohortIDs []int64 `json:"cohort_ids" example:"1,2"`
}
