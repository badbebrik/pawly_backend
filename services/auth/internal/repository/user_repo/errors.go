package user_repo

import "errors"

var (
	ErrUserNotFound = errors.New("user_not_found")
	ErrEmailTaken   = errors.New("email_already_taken")
)
