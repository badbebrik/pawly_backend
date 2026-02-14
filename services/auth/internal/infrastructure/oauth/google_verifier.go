package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidToken        = errors.New("invalid_token")
	ErrProviderUnavailable = errors.New("provider_unavailable")
)

type Claims struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
}

type Verifier interface {
	VerifyGoogleIDToken(ctx context.Context, idToken string) (*Claims, error)
}

type GoogleVerifier struct {
	client   *http.Client
	clientID string
}

type tokenInfoResponse struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Exp           string `json:"exp"`
}

func NewGoogleVerifier(timeout time.Duration, clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		client:   &http.Client{Timeout: timeout},
		clientID: clientID,
	}
}

func (v *GoogleVerifier) VerifyGoogleIDToken(ctx context.Context, idToken string) (*Claims, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, ErrInvalidToken
	}

	reqURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google tokeninfo request: %w", ErrProviderUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 500 {
			return nil, ErrProviderUnavailable
		}
		return nil, ErrInvalidToken
	}

	var out tokenInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode tokeninfo response: %w", err)
	}

	if out.Sub == "" {
		return nil, ErrInvalidToken
	}
	if v.clientID != "" && out.Aud != v.clientID {
		return nil, ErrInvalidToken
	}
	if out.Exp != "" {
		exp, err := strconv.ParseInt(out.Exp, 10, 64)
		if err != nil || exp < time.Now().Unix() {
			return nil, ErrInvalidToken
		}
	}

	return &Claims{
		Provider:      "google",
		Subject:       out.Sub,
		Email:         strings.TrimSpace(strings.ToLower(out.Email)),
		EmailVerified: parseBool(out.EmailVerified),
	}, nil
}

func parseBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || t == "1"
	default:
		return false
	}
}
