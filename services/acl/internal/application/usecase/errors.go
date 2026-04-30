package usecase

import "errors"

var (
	ErrForbidden      = errors.New("forbidden")
	ErrNotFound       = errors.New("not_found")
	ErrInvalidInput   = errors.New("invalid_input")
	ErrConflict       = errors.New("conflict")
	ErrNotImplemented = errors.New("not_implemented")
)
