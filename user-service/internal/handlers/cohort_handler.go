package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"user-service/internal/dto"
	"user-service/internal/middleware"
	"user-service/internal/models"
	"user-service/internal/security"

	"github.com/google/uuid"
)

const teacherRole = "teacher"

type CohortService interface {
	CreateCohort(ctx context.Context, name string, ownerID uuid.UUID) (*models.Cohort, error)
	GetCohorts(ctx context.Context, userID uuid.UUID) ([]*models.Cohort, error)
	GetCohortWithUsers(ctx context.Context, cohortID int64) (*models.CohortWithUsers, error)
	AddsUserToCohortByInvite(ctx context.Context, cohortID int64, userID uuid.UUID) error
	GenerateInviteTokenToCohort(ctx context.Context, cohortID int64) (string, error)
	IsCohortOwnedByUser(ctx context.Context, cohortID int64, userID uuid.UUID) (bool, error)
	IsUserInCohorts(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]int64, error)
}

type InviteTokenParser interface {
	ParseInviteToken(tokenStr string) (*security.InviteClaims, error)
}

type CohortHandler struct {
	cohortService     CohortService
	inviteTokenParser InviteTokenParser
}

func NewCohortHandler(cohortService CohortService, inviteTokenParser InviteTokenParser) *CohortHandler {
	return &CohortHandler{cohortService: cohortService, inviteTokenParser: inviteTokenParser}
}

// CreateCohort creates a new cohort for the authenticated teacher.
// @Summary Create a new cohort
// @Description Creates a new cohort owned by the authenticated teacher. Requires teacher role.
// @Tags cohorts
// @Security Bearer []
// @Param request body dto.CohortCreateRequest true "Cohort creation request"
// @Success 200 {object} dto.CohortResponse "Cohort created successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid input or user ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - teacher role required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /cohorts [post]
func (h *CohortHandler) CreateCohort(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if !strings.EqualFold(claims.Role, teacherRole) {
		writeError(w, http.StatusForbidden, "forbidden: teacher role required")
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req dto.CohortCreateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid req body")
		return
	}

	cohort, err := h.cohortService.CreateCohort(r.Context(), req.Name, userID)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create cohort")
		return
	}

	resp := dto.CohortResponse{
		ID:   strconv.FormatInt(cohort.ID, 10),
		Name: cohort.Name,
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetCohorts retrieves all cohorts for the authenticated user.
// @Summary Get user cohorts
// @Description Retrieves all cohorts that the authenticated user is a member of or owns
// @Tags cohorts
// @Security Bearer []
// @Success 200 {array} dto.CohortResponse "List of cohorts retrieved successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid user ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /cohorts [get]
func (h *CohortHandler) GetCohorts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	cohorts, err := h.cohortService.GetCohorts(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get cohorts")
		return
	}

	resp := make([]dto.CohortResponse, 0, len(cohorts))
	for _, cohort := range cohorts {
		resp = append(resp, dto.CohortResponse{
			ID:   strconv.FormatInt(cohort.ID, 10),
			Name: cohort.Name,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetCohortMembers retrieves all members of a specific cohort.
// @Summary Get cohort members
// @Description Retrieves all members of a cohort. User must be the cohort owner or a member.
// @Tags cohorts
// @Security Bearer []
// @Param id path string true "Cohort ID"
// @Success 200 {object} dto.CohortWithUsersResponse "Cohort with members retrieved successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid cohort ID or user ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user is not a member or owner"
// @Failure 404 {object} map[string]string "Not found - cohort does not exist"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /cohorts/{id}/members [get]
func (h *CohortHandler) GetCohortMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	cohortID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort id")
		return
	}

	cohort, err := h.cohortService.GetCohortWithUsers(r.Context(), cohortID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			writeError(w, http.StatusNotFound, "cohort not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to get cohort")
		return
	}

	canView := cohort.OwnerID == userID
	if !canView {
		for _, member := range cohort.Users {
			if member.ID == userID {
				canView = true
				break
			}
		}
	}

	if !canView {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	users := make([]dto.UserResponse, 0, len(cohort.Users))
	for _, user := range cohort.Users {
		users = append(users, dto.UserResponse{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Preferences: user.Preferences,
		})
	}

	resp := dto.CohortWithUsersResponse{
		ID:    strconv.FormatInt(cohort.ID, 10),
		Name:  cohort.Name,
		Users: users,
	}

	writeJSON(w, http.StatusOK, resp)
}

// JoinCohort allows a user to join a cohort using an invite token.
// @Summary Join a cohort
// @Description Joins a cohort using a valid invite token. Teachers cannot join cohorts.
// @Tags cohorts
// @Security Bearer []
// @Param request body dto.CohortJoinRequest true "Join cohort request with invite token"
// @Success 200 {object} map[string]string "Cohort joined successfully, returns cohort_id"
// @Failure 400 {object} map[string]string "Bad request - invalid token or user ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - teachers cannot join cohorts"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /cohorts/join [post]
func (h *CohortHandler) JoinCohort(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if strings.EqualFold(claims.Role, teacherRole) {
		writeError(w, http.StatusForbidden, "forbidden: teachers cannot join cohorts")
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if h.inviteTokenParser == nil {
		writeError(w, http.StatusInternalServerError, "invite token parser is not configured")
		return
	}

	var req dto.CohortJoinRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid req body")
		return
	}

	inviteClaims, err := h.inviteTokenParser.ParseInviteToken(req.Token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invite token")
		return
	}

	cohortID, err := strconv.ParseInt(inviteClaims.CohortID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invite token")
		return
	}

	if err := h.cohortService.AddsUserToCohortByInvite(r.Context(), cohortID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to join cohort")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cohort_id": strconv.FormatInt(cohortID, 10)})
}

// GenerateInviteToken generates an invite token for a cohort.
// @Summary Generate invite token
// @Description Generates a new invite token for a cohort. Only the cohort owner can generate tokens.
// @Tags cohorts
// @Security Bearer []
// @Param id path string true "Cohort ID"
// @Success 200 {object} map[string]string "Invite token generated successfully, returns token"
// @Failure 400 {object} map[string]string "Bad request - invalid cohort ID or user ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - only cohort owner can generate tokens"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /cohorts/{id}/invite [post]
func (h *CohortHandler) GenerateInviteToken(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	cohortID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort id")
		return
	}

	isOwner, err := h.cohortService.IsCohortOwnedByUser(r.Context(), cohortID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check ownership")
		return
	}

	if !isOwner {
		writeError(w, http.StatusForbidden, "forbidden: only cohort owner can generate invite token")
		return
	}

	token, err := h.cohortService.GenerateInviteTokenToCohort(r.Context(), cohortID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate invite token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// IsOwner checks if a user owns a cohort (internal endpoint).
// @Summary Check cohort ownership
// @Description Internal endpoint to check if a user owns a specific cohort. Requires internal token authentication.
// @Tags cohorts
// @Security InternalToken []
// @Param request body dto.CohortIsOwnedRequest true "Ownership check request"
// @Success 200 {object} map[string]bool "Ownership status returned as is_owner boolean"
// @Failure 400 {object} map[string]string "Bad request - invalid cohort ID or user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /internal/cohorts/can-edit [post]
func (h *CohortHandler) IsOwner(w http.ResponseWriter, r *http.Request) {
	var req dto.CohortIsOwnedRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid req body")
		return
	}

	cohortID, err := strconv.ParseInt(req.CohortID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort id")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	isOwner, err := h.cohortService.IsCohortOwnedByUser(r.Context(), cohortID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check ownership")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"is_owner": isOwner})
}

// IsUserIn checks if a user is a member of any specified cohorts (internal endpoint).
// @Summary Check user cohort membership
// @Description Internal endpoint to check which of the specified cohorts a user is a member of. Requires internal token authentication.
// @Tags cohorts
// @Security InternalToken []
// @Param request body dto.CohortIsUserInRequest true "Membership check request with user ID and cohort IDs"
// @Success 200 {object} dto.CohortIsUserInResponse "List of cohort IDs the user is a member of"
// @Failure 400 {object} map[string]string "Bad request - invalid user ID or cohort IDs"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /internal/cohorts/is-user-in [post]
func (h *CohortHandler) IsUserIn(w http.ResponseWriter, r *http.Request) {
	var req dto.CohortIsUserInRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid req body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	cohortIDs, err := h.cohortService.IsUserInCohorts(r.Context(), userID, req.CohortIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check memberships")
		return
	}

	writeJSON(w, http.StatusOK, dto.CohortIsUserInResponse{CohortIDs: cohortIDs})
}
