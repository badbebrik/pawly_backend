package aclinternalclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"health/internal/application/ports"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("acl http base url is empty")
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("internal service token is empty")
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (c *Client) ListPetUserIDs(ctx context.Context, petID uuid.UUID) ([]uuid.UUID, error) {
	body, err := json.Marshal(map[string]any{"pet_id": petID.String()})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/acl:list-members-for-pet", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acl internal list members for pet: status %d", resp.StatusCode)
	}

	var payload struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	items := make([]uuid.UUID, 0, len(payload.UserIDs))
	seen := make(map[uuid.UUID]struct{}, len(payload.UserIDs))
	for _, raw := range payload.UserIDs {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil || id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, id)
	}
	return items, nil
}

var _ ports.PetUserLister = (*Client)(nil)
