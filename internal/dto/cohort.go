package dto

type CohortResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CohortWithUsersResponse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Users []UserResponse `json:"users"`
}

type CohortCreateRequest struct {
	Name string `json:"name"`
}

type CohortJoinRequest struct {
	Token string `json:"token"`
}

type CohortIsOwnedRequest struct {
	CohortID string `json:"cohort_id"`
	UserID   string `json:"user_id"`
}
