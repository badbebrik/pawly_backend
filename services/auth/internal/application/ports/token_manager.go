package ports

import (
	"time"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenTypeReset   = "password_reset"
)

type TokenClaims struct {
	Subject   string
	SessionID string
	Type      string
	ExpiresAt time.Time
	Email     string
}

type TokenManager interface {
	GenerateAccessToken(userID, sessionID string) (string, error)
	GenerateRefreshToken(userID, sessionID string) (string, error)
	GeneratePasswordResetToken(userID, email string) (string, error)
	ValidateToken(tokenStr string) (*TokenClaims, error)
	EnsureTokenType(claims *TokenClaims, tokenType string) error
}
