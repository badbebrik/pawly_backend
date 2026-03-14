package service

import "errors"

var (
	ErrInvalidInput    = errors.New("invalid_input")
	ErrInvalidLocale   = errors.New("invalid_locale")
	ErrInvalidTimezone = errors.New("invalid_timezone")
)
