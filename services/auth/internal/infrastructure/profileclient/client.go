package profileclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client interface {
	CreateProfile(ctx context.Context, userID uuid.UUID, locale string) error
}

type HTTPClient struct {
	baseURL       string
	internalToken string
	httpClient    *http.Client
}

type createProfileRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Locale *string   `json:"locale,omitempty"`
}

func New(baseURL, internalToken string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL:       strings.TrimRight(baseURL, "/"),
		internalToken: internalToken,
		httpClient:    &http.Client{Timeout: timeout},
	}
}

func (c *HTTPClient) CreateProfile(ctx context.Context, userID uuid.UUID, locale string) error {
	var localePtr *string
	if strings.TrimSpace(locale) != "" {
		l := locale
		localePtr = &l
	}

	payload := createProfileRequest{UserID: userID, Locale: localePtr}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := c.baseURL + "/internal/v1/profile/users"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.internalToken != "" {
		req.Header.Set("X-Internal-Token", c.internalToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("profile service returned status %d", resp.StatusCode)
}
