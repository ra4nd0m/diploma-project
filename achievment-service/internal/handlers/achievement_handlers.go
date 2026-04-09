package handlers

import (
	"achievement-service/internal/middleware"
	"achievement-service/internal/services"
	achievement_creation "achievement-service/internal/services/achievement_creation"
	achievementissue "achievement-service/internal/services/achievement_issue"
	achievement_reading_service "achievement-service/internal/services/achievement_reading"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AchievementCreationService interface {
	CreateAchievement(ctx context.Context, input achievement_creation.Input) (int64, error)
}

type AchievementIssueService interface {
	IssueAchievement(ctx context.Context, input achievementissue.Input) (*achievementissue.Output, error)
}

type AchievementReadingService interface {
	GetVisibleAchievements(ctx context.Context, userID uuid.UUID, cohortIDs []int64) ([]*achievement_reading_service.Output, error)
	GetOwnedAchievements(ctx context.Context, ownerID uuid.UUID, cohortIDs []int64) ([]*achievement_reading_service.Output, error)
	GetRecipientAchievements(ctx context.Context, requestUserID, recipientID uuid.UUID, cohortIDs []int64) ([]*achievement_reading_service.Output, error)
}

type AchievementHandler struct {
	creationService AchievementCreationService
	issueService    AchievementIssueService
	readingService  AchievementReadingService
}

func NewAchievementHandler(
	creationService AchievementCreationService,
	issueService AchievementIssueService,
	readingService AchievementReadingService,
) *AchievementHandler {
	return &AchievementHandler{
		creationService: creationService,
		issueService:    issueService,
		readingService:  readingService,
	}
}

// Achievements godoc
// @Summary List or create achievements
// @Description Handles both POST (create new achievement) and GET (list visible achievements) requests
// @Tags achievements
// @Accept json
// @Produce json
// @Security Bearer
func (h *AchievementHandler) Achievements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateAchievement(w, r)
	case http.MethodGet:
		h.GetAchievements(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodPost, http.MethodGet)
	}
}

// CreateAchievement godoc
// @Summary Create a new achievement
// @Description Creates a new achievement in the system. Requires authentication.
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body createAchievementRequestDTO true "Achievement creation request"
// @Success 201 {object} createAchievementResponseDTO
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /achievements [post]
func (h *AchievementHandler) CreateAchievement(w http.ResponseWriter, r *http.Request) {
	var req createAchievementRequestDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID, err := userIDFromClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input := req.toInput(userID)

	id, err := h.creationService.CreateAchievement(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createAchievementResponseDTO{ID: id})
}

// GetAchievements godoc
// @Summary List visible achievements
// @Description Retrieves all achievements visible to the authenticated user from specified cohorts
// @Tags achievements
// @Accept json
// @Produce json
// @Param cohort_ids query string false "Comma-separated list of cohort IDs"
// @Success 200 {array} achievementResponseDTO
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /achievements [get]
func (h *AchievementHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cohortIDs, err := parseCohortIDs(r.URL.Query().Get("cohort_ids"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort_ids")
		return
	}

	items, err := h.readingService.GetVisibleAchievements(r.Context(), userID, cohortIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, achievementsResponseFromOutput(items))
}

// GetOwnedAchievements godoc
// @Summary Get achievements owned by the user
// @Description Retrieves all achievements created/owned by the authenticated user from specified cohorts
// @Tags achievements
// @Accept json
// @Produce json
// @Param cohort_ids query string false "Comma-separated list of cohort IDs"
// @Success 200 {array} achievementResponseDTO
// @Failure 400 {object} map[string]string "Invalid query parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /achievements/owned [get]
func (h *AchievementHandler) GetOwnedAchievements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	ownerID, err := userIDFromClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cohortIDs, err := parseCohortIDs(r.URL.Query().Get("cohort_ids"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort_ids")
		return
	}

	items, err := h.readingService.GetOwnedAchievements(r.Context(), ownerID, cohortIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, achievementsResponseFromOutput(items))
}

// GetRecipientAchievements godoc
// @Summary Get achievements of a specific recipient
// @Description Retrieves all achievements issued to a specific user from specified cohorts. Requires authentication.
// @Tags achievements
// @Accept json
// @Produce json
// @Param recipientID path string true "UUID of the achievement recipient"
// @Param cohort_ids query string false "Comma-separated list of cohort IDs"
// @Success 200 {array} achievementResponseDTO
// @Failure 400 {object} map[string]string "Invalid recipient ID or query parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /achievements/recipient/{recipientID} [get]
func (h *AchievementHandler) GetRecipientAchievements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	requestUserID, err := userIDFromClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	recipientID, err := uuid.Parse(chi.URLParam(r, "recipientID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recipient id")
		return
	}

	cohortIDs, err := parseCohortIDs(r.URL.Query().Get("cohort_ids"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cohort_ids")
		return
	}

	items, err := h.readingService.GetRecipientAchievements(r.Context(), requestUserID, recipientID, cohortIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, achievementsResponseFromOutput(items))
}

// IssueAchievement godoc
// @Summary Issue an achievement to a recipient
// @Description Issues (awards) an achievement to a specific user. Requires authentication and proper authorization.
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body issueAchievementRequestDTO true "Achievement issuance request"
// @Success 201 {object} issueAchievementResponseDTO
// @Failure 400 {object} map[string]string "Invalid request body or recipient ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security Bearer
// @Router /achievements/issue [post]
func (h *AchievementHandler) IssueAchievement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req issueAchievementRequestDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	issuerID, err := userIDFromClaims(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	input, err := req.toInput(issuerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid recipient_id")
		return
	}

	out, err := h.issueService.IssueAchievement(r.Context(), input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, issueAchievementResponseFromOutput(out))
}

func decodeJSON(r *http.Request, out any) error {
	defer func() {
		_ = r.Body.Close()
	}()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidInput), errors.Is(err, services.ErrInvalidCondition):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, services.ErrNotFound),
		errors.Is(err, services.ErrAchievementNotFound),
		errors.Is(err, services.ErrConditionTypeNotFound),
		errors.Is(err, services.ErrIssuanceKindNotFound),
		errors.Is(err, services.ErrAccessModeNotFound),
		errors.Is(err, services.ErrStatusNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrAlreadyIssued):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeMethodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func userIDFromClaims(ctx context.Context) (uuid.UUID, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return uuid.Nil, errors.New("missing claims")
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func parseCohortIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int64{}, nil
	}

	parts := strings.Split(raw, ",")
	cohortIDs := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("invalid cohort id")
		}
		cohortIDs = append(cohortIDs, id)
	}

	return cohortIDs, nil
}
