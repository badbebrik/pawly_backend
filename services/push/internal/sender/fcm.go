package sender

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"push/internal/domain/model"
)

const (
	fcmScope    = "https://www.googleapis.com/auth/firebase.messaging"
	fcmTokenURL = "https://oauth2.googleapis.com/token"
	fcmAudience = fcmTokenURL
	tokenLeeway = 30 * time.Second
	sendTimeout = 10 * time.Second
)

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

type FCMSender struct {
	projectID string
	account   serviceAccount
	key       *rsa.PrivateKey
	http      *http.Client

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func NewFCMSender(projectID, credentialsFile string) (*FCMSender, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("fcm project id is empty")
	}
	if strings.TrimSpace(credentialsFile) == "" {
		return nil, fmt.Errorf("fcm credentials file is empty")
	}

	raw, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	var account serviceAccount
	if err := json.Unmarshal(raw, &account); err != nil {
		return nil, err
	}
	if strings.TrimSpace(account.ClientEmail) == "" || strings.TrimSpace(account.PrivateKey) == "" {
		return nil, fmt.Errorf("invalid fcm service account json")
	}
	if strings.TrimSpace(account.TokenURI) == "" {
		account.TokenURI = fcmTokenURL
	}

	key, err := parsePrivateKey(account.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &FCMSender{
		projectID: projectID,
		account:   account,
		key:       key,
		http: &http.Client{
			Timeout: sendTimeout,
		},
	}, nil
}

func (s *FCMSender) Send(ctx context.Context, device model.DeviceToken, msg model.PushMessage) error {
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return err
	}

	body := fcmSendRequest{
		Message: fcmMessage{
			Token: device.PushToken,
			Notification: fcmNotification{
				Title: msg.Title,
				Body:  msg.Body,
			},
			Data: msg.Data,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.sendURL(), strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var fcmErr fcmErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&fcmErr); err != nil {
		return fmt.Errorf("fcm send status %d", resp.StatusCode)
	}
	if isInvalidTokenError(fcmErr) {
		return ErrInvalidDeviceToken
	}

	return fmt.Errorf("fcm send failed: status=%d message=%s", resp.StatusCode, fcmErr.Error.Message)
}

func (s *FCMSender) sendURL() string {
	return fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.projectID)
}

func (s *FCMSender) getAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.accessToken != "" && time.Now().Before(s.expiresAt.Add(-tokenLeeway)) {
		token := s.accessToken
		s.mu.Unlock()
		return token, nil
	}
	s.mu.Unlock()

	assertion, err := s.buildJWT()
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.account.TokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("google oauth token status %d", resp.StatusCode)
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", fmt.Errorf("empty google oauth access token")
	}

	expiresAt := time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	s.mu.Lock()
	s.accessToken = payload.AccessToken
	s.expiresAt = expiresAt
	s.mu.Unlock()
	return payload.AccessToken, nil
}

func (s *FCMSender) buildJWT() (string, error) {
	headerJSON, err := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	claimsJSON, err := json.Marshal(map[string]any{
		"iss":   s.account.ClientEmail,
		"scope": fcmScope,
		"aud":   fcmAudience,
		"iat":   now,
		"exp":   now + 3600,
	})
	if err != nil {
		return "", err
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func parsePrivateKey(raw string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("invalid pem private key")
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not rsa")
		}
		return rsaKey, nil
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func isInvalidTokenError(resp fcmErrorResponse) bool {
	if strings.EqualFold(strings.TrimSpace(resp.Error.Status), "INVALID_ARGUMENT") {
		for _, d := range resp.Error.Details {
			if strings.Contains(d.ErrorCode, "UNREGISTERED") || strings.Contains(d.ErrorCode, "INVALID_ARGUMENT") {
				return true
			}
		}
	}
	for _, d := range resp.Error.Details {
		if strings.Contains(d.ErrorCode, "UNREGISTERED") {
			return true
		}
	}
	return false
}

type fcmSendRequest struct {
	Message fcmMessage `json:"message"`
}

type fcmMessage struct {
	Token        string            `json:"token"`
	Notification fcmNotification   `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
}

type fcmNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Details []struct {
			Type      string `json:"@type"`
			ErrorCode string `json:"errorCode"`
		} `json:"details"`
	} `json:"error"`
}

var _ Sender = (*FCMSender)(nil)
var _ Sender = (*NoopSender)(nil)
