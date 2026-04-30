package ports

import "errors"

var (
	ErrInvalidInput = errors.New("invalid_input")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not_found")
	ErrConflict     = errors.New("conflict")
)
