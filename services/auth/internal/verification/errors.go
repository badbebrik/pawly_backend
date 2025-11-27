package verification

import "errors"

var (
	ErrCodeNotFound    = errors.New("verification_code_not_found")
	ErrCodeExpired     = errors.New("verification_code_expired")
	ErrCodeInvalid     = errors.New("invalid_verification_code")
	ErrTooManyAttempts = errors.New("too_many_attempts")
	ErrResendTooSoon   = errors.New("resend_not_available_yet")
)
