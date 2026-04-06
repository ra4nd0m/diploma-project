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

func (h *CohortHandler) GetCohortMembers(w http.ResponseWriter, r *http.Request) {
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

func (h *CohortHandler) JoinCohort(w http.ResponseWriter, r *http.Request) {
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

func (h *CohortHandler) GenerateInviteToken(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.ClaimsFromContext(r.Context()); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cohortID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort id")
		return
	}

	token, err := h.cohortService.GenerateInviteTokenToCohort(r.Context(), cohortID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate invite token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

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
