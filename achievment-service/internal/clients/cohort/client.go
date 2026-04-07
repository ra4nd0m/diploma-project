package cohortclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

func NewClient(baseURL, internalToken string) *Client {
	return &Client{
		baseURL:       strings.TrimRight(baseURL, "/"),
		internalToken: internalToken,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type canEditRequest struct {
	UserID   string `json:"user_id"`
	CohortID string `json:"cohort_id"`
}

type canEditResponse struct {
	IsOwner bool `json:"is_owner"`
}

func (c *Client) CanEditCohort(
	ctx context.Context,
	userID uuid.UUID,
	cohortID int64,
) (bool, error) {
	reqBody := canEditRequest{
		UserID:   userID.String(),
		CohortID: strconv.FormatInt(cohortID, 10),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("marshal cohort auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/internal/cohorts/can-edit",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return false, fmt.Errorf("build cohort auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.internalToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("send cohort auth request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("cohort auth returned status %d", resp.StatusCode)
	}

	var out canEditResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, fmt.Errorf("decode cohort auth response: %w", err)
	}

	return out.IsOwner, nil
}

func (c *Client) IsUserInCohort(
	ctt context.Context,
	userID uuid.UUID,
	cohortIDs []int64,
) ([]int64, error) {
	reqBody := struct {
		UserID    uuid.UUID `json:"user_id"`
		CohortIDs []int64   `json:"cohort_ids"`
	}{
		UserID:    userID,
		CohortIDs: cohortIDs,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal cohort membership request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctt,
		http.MethodPost,
		c.baseURL+"/internal/cohorts/is-user-in",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("build cohort membership request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.internalToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send cohort membership request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cohort membership returned status %d", resp.StatusCode)
	}

	var out struct {
		CohortIDs []int64 `json:"cohort_ids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode cohort membership response: %w", err)
	}

	return out.CohortIDs, nil
}
