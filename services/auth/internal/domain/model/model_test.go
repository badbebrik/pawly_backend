package model

import (
	"errors"
	"testing"
	"time"
)

func TestUserCanLoginWithPassword(t *testing.T) {
	hash := "hashed"

	tests := []struct {
		name    string
		user    User
		wantErr error
	}{
		{
			name: "ok",
			user: User{
				PasswordHash: &hash,
				IsVerified:   true,
				IsActive:     true,
			},
		},
		{
			name: "inactive",
			user: User{
				PasswordHash: &hash,
				IsVerified:   true,
				IsActive:     false,
			},
			wantErr: ErrUserInactive,
		},
		{
			name: "unverified",
			user: User{
				PasswordHash: &hash,
				IsVerified:   false,
				IsActive:     true,
			},
			wantErr: ErrUserUnverified,
		},
		{
			name: "no password account",
			user: User{
				IsVerified: true,
				IsActive:   true,
			},
			wantErr: ErrPasswordAuthUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.CanLoginWithPassword()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanLoginWithPassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSessionCanRefresh(t *testing.T) {
	now := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		session Session
		hash    string
		wantErr error
	}{
		{
			name: "ok",
			session: Session{
				RefreshTokenHash: "hash",
				ExpiresAt:        now.Add(time.Hour),
			},
			hash: "hash",
		},
		{
			name: "revoked",
			session: Session{
				RefreshTokenHash: "hash",
				ExpiresAt:        now.Add(time.Hour),
				IsRevoked:        true,
			},
			hash:    "hash",
			wantErr: ErrSessionRevoked,
		},
		{
			name: "expired",
			session: Session{
				RefreshTokenHash: "hash",
				ExpiresAt:        now.Add(-time.Second),
			},
			hash:    "hash",
			wantErr: ErrSessionExpired,
		},
		{
			name: "mismatch",
			session: Session{
				RefreshTokenHash: "hash",
				ExpiresAt:        now.Add(time.Hour),
			},
			hash:    "other",
			wantErr: ErrRefreshTokenMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.CanRefresh(now, tt.hash)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CanRefresh() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
