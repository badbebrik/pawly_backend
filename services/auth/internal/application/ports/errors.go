package ports

import "errors"

var (
	ErrNotFound                 = errors.New("not_found")
	ErrConflict                 = errors.New("conflict")
	ErrCodeInvalid              = errors.New("code_invalid")
	ErrCodeNotFound             = errors.New("code_not_found")
	ErrCodeExpired              = errors.New("code_expired")
	ErrTooManyAttempts          = errors.New("too_many_attempts")
	ErrResendTooSoon            = errors.New("resend_too_soon")
	ErrOAuthInvalidToken        = errors.New("oauth_invalid_token")
	ErrOAuthProviderUnavailable = errors.New("oauth_provider_unavailable")
)
