package usecase

import "errors"

var (
	ErrInvalidEmailOrPassword = errors.New("invalid_email_or_password")
	ErrIncorrectFormat        = errors.New("incorrect_format")
	ErrWeakPassword           = errors.New("weak_password")

	ErrEmailAlreadyTaken    = errors.New("email_taken")
	ErrEmailAlreadyVerified = errors.New("email_already_verified")
	ErrVerificationFailed   = errors.New("verification_failed")
	ErrCannotResendYet      = errors.New("cannot_resend_yet")

	ErrUserNotFound     = errors.New("user_not_found")
	ErrSessionNotFound  = errors.New("session_not_found")
	ErrSessionExpired   = errors.New("session_expired")
	ErrSessionRevoked   = errors.New("session_revoked")
	ErrRefreshMismatch  = errors.New("refresh_token_mismatch")
	ErrUserBlocked      = errors.New("user_blocked")
	ErrEmailNotVerified = errors.New("email_not_verified")

	ErrVerificationCodeInvalid  = errors.New("verification_code_invalid")
	ErrVerificationCodeExpired  = errors.New("verification_code_expired")
	ErrVerificationTooMany      = errors.New("verification_too_many_attempts")
	ErrProfileCreationFailed    = errors.New("profile_creation_failed")
	ErrOAuthInvalidToken        = errors.New("oauth_invalid_token")
	ErrOAuthProviderUnavailable = errors.New("oauth_provider_unavailable")

	ErrUnauthorized = errors.New("unauthorized")
)
