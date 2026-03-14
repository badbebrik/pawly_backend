package model

import "errors"

var (
	ErrUserInactive            = errors.New("user_inactive")
	ErrUserUnverified          = errors.New("user_unverified")
	ErrPasswordAuthUnavailable = errors.New("password_auth_unavailable")

	ErrSessionExpired       = errors.New("session_expired")
	ErrSessionRevoked       = errors.New("session_revoked")
	ErrRefreshTokenMismatch = errors.New("refresh_token_mismatch")
)
