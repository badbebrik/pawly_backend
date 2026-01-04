package service

import "errors"

var (
	ErrInvalidEmailOrPassword = errors.New("invalid_email_or_password")
	ErrIncorrectFormat        = errors.New("incorrect_format")
	ErrWeakPassword           = errors.New("weak_password")

	ErrEmailAlreadyTaken   = errors.New("email_taken")
	ErrProfileCreateFailed = errors.New("profile_create_failed")
	ErrVerificationFailed  = errors.New("verification_failed")
	ErrCannotResendYet     = errors.New("cannot_resend_yet")

	ErrUserNotFound     = errors.New("user_not_found")
	ErrSessionNotFound  = errors.New("session_not_found")
	ErrSessionExpired   = errors.New("session_expired")
	ErrUserBlocked      = errors.New("user_blocked")
	ErrEmailNotVerified = errors.New("email_not_verified")
)
