package session_repo

import "errors"

var (
	ErrSessionNotFound = errors.New("session_not_found")
	ErrSessionExpired  = errors.New("session_expired")
)
