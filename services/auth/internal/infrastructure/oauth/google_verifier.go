package oauth

import (
	"auth/internal/application/ports"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func NewGoogleVerifier(timeout time.Duration, clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		client:   &http.Client{Timeout: timeout},
		clientID: clientID,
	}
}

func (v *GoogleVerifier) VerifyGoogleIDToken(ctx context.Context, idToken string) (*ports.OAuthClaims, error) {
	if strings.TrimSpace(idToken) == "" {
		return nil, ports.ErrOAuthInvalidToken
	}

	reqURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google tokeninfo request: %w", ports.ErrOAuthProviderUnavailable)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 500 {
			return nil, ports.ErrOAuthProviderUnavailable
		}
		return nil, ports.ErrOAuthInvalidToken
	}

	var out tokenInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode tokeninfo response: %w", err)
	}

	if out.Sub == "" {
		return nil, ports.ErrOAuthInvalidToken
	}
	if v.clientID != "" && out.Aud != v.clientID {
		return nil, ports.ErrOAuthInvalidToken
	}
	if out.Exp != "" {
		exp, err := strconv.ParseInt(out.Exp, 10, 64)
		if err != nil || exp < time.Now().Unix() {
			return nil, ports.ErrOAuthInvalidToken
		}
	}

	return &ports.OAuthClaims{
		Provider:      "google",
		Subject:       out.Sub,
		Email:         strings.TrimSpace(strings.ToLower(out.Email)),
		EmailVerified: parseBool(out.EmailVerified),
		FirstName:     strings.TrimSpace(out.GivenName),
		LastName:      strings.TrimSpace(out.FamilyName),
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
