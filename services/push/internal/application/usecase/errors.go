package usecase

import "errors"

var (
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid_input")
)
