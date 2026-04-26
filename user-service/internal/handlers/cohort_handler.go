// Package handlers contains HTTP request handlers for the User Service API.
//
// This package implements the HTTP handlers for user and cohort-related endpoints.
// Handlers are organized by domain:
//   - cohort_handler.go: Cohort management and membership operations
//   - user_handler.go: User profile operations
//   - respond.go: Helper functions for HTTP responses
//
// All handlers use middleware-provided claims for authentication and authorization.
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

// CohortService defines the interface for cohort business logic operations.
type CohortService interface {
	CreateCohort(ctx context.Context, name string, ownerID uuid.UUID) (*models.Cohort, error)
	GetCohorts(ctx context.Context, userID uuid.UUID) ([]*models.Cohort, error)
	GetCohortWithUsers(ctx context.Context, cohortID int64) (*models.CohortWithUsers, error)
	AddsUserToCohortByInvite(ctx context.Context, cohortID int64, userID uuid.UUID) error
	RemoveUserFromCohort(ctx context.Context, cohortID int64, userID uuid.UUID) error
	GenerateInviteTokenToCohort(ctx context.Context, cohortID int64) (string, error)
	IsCohortOwnedByUser(ctx context.Context, cohortID int64, userID uuid.UUID) (bool, error)
	IsUserInCohorts(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]int64, error)
}

// InviteTokenParser defines the interface for parsing and validating invite tokens.
type InviteTokenParser interface {
	ParseInviteToken(tokenStr string) (*security.InviteClaims, error)
}

// CohortHandler handles HTTP requests related to cohort operations including creation,
// listing, membership management, and invite token generation.
// It requires a CohortService for business logic and an InviteTokenParser for token validation.
type CohortHandler struct {
	cohortService     CohortService
	inviteTokenParser InviteTokenParser
}

// NewCohortHandler creates a new CohortHandler with the provided service dependencies.
func NewCohortHandler(cohortService CohortService, inviteTokenParser InviteTokenParser) *CohortHandler {
	return &CohortHandler{cohortService: cohortService, inviteTokenParser: inviteTokenParser}
}

// CreateCohort creates a cohort for the authenticated teacher.
// @Summary Create cohort
// @Description Creates a new cohort owned by the authenticated teacher.
// @Tags cohorts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req body dto.CohortCreateRequest true "Create cohort request"
// @Success 200 {object} dto.CohortResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// GetCohorts returns cohorts visible to the authenticated user.
// @Summary List cohorts
// @Description Returns the list of cohorts available to the authenticated user.
// @Tags cohorts
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.CohortResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// GetCohortMembers returns cohort members if the user has access.
// @Summary Get cohort members
// @Description Returns cohort details and members if the authenticated user is the owner or a member.
// @Tags cohorts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cohort ID"
// @Success 200 {object} dto.CohortWithUsersResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// RemoveUserFromCohort removes a user from a cohort.
// @Summary Remove user from cohort
// @Description Removes a user from a cohort. Allowed for the user themselves or the cohort owner.
// @Tags cohorts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cohort ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cohorts/{id}/members/{user_id} [delete]
func (h *CohortHandler) RemoveUserFromCohort(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	actorUserID, err := uuid.Parse(claims.Sub)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	cohortID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort id")
		return
	}

	targetUserID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	canRemove := actorUserID == targetUserID
	if !canRemove {
		isOwner, err := h.cohortService.IsCohortOwnedByUser(r.Context(), cohortID, actorUserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to check ownership")
			return
		}

		if !isOwner {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
	}

	if err := h.cohortService.RemoveUserFromCohort(r.Context(), cohortID, targetUserID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove user from cohort")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// JoinCohort joins the authenticated student to a cohort using an invite token.
// @Summary Join cohort
// @Description Joins the authenticated user to a cohort identified by an invite token.
// @Tags cohorts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req body dto.CohortJoinRequest true "Join cohort request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// GenerateInviteToken creates an invite token for a cohort owned by the user.
// @Summary Generate invite token
// @Description Generates an invite token for a cohort owned by the authenticated user.
// @Tags cohorts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cohort ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// IsOwner checks whether a user owns a cohort.
// @Summary Check cohort ownership
// @Description Checks whether the supplied user owns the supplied cohort.
// @Tags internal
// @Accept json
// @Produce json
// @Security InternalToken
// @Param req body dto.CohortIsOwnedRequest true "Ownership check request"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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

// IsUserIn checks whether a user belongs to any of the supplied cohorts.
// @Summary Check cohort membership
// @Description Checks whether the supplied user belongs to any of the supplied cohorts.
// @Tags internal
// @Accept json
// @Produce json
// @Security InternalToken
// @Param req body dto.CohortIsUserInRequest true "Membership check request"
// @Success 200 {object} dto.CohortIsUserInResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
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
