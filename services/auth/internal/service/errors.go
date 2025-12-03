package auth

import "errors"

var (
	ErrIncorrectFormat     = errors.New("incorrect_format")
	ErrEmailAlreadyTaken   = errors.New("email_taken")
	ErrProfileCreateFailed = errors.New("profile_create_failed")
	ErrVerificationFailed  = errors.New("verification_code_create_failed")
)
