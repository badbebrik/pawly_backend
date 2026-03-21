package usecase

import "errors"

var (
	ErrInvalidInput = errors.New("invalid_input")
	ErrForbidden    = errors.New("forbidden")
)
