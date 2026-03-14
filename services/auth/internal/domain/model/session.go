package model

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
	ExpiresAt        time.Time `json:"expires_at"`
	IsRevoked        bool      `json:"is_revoked"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *Session) MatchesRefreshToken(hash string) bool {
	return s.RefreshTokenHash == hash
}

func (s *Session) CanRefresh(now time.Time, refreshTokenHash string) error {
	if s.IsRevoked {
		return ErrSessionRevoked
	}
	if s.IsExpired(now) {
		return ErrSessionExpired
	}
	if !s.MatchesRefreshToken(refreshTokenHash) {
		return ErrRefreshTokenMismatch
	}
	return nil
}
