package usecase

import "errors"

var (
	ErrInvalidInput   = errors.New("invalid_input")
	ErrForbidden      = errors.New("forbidden")
	ErrNotImplemented = errors.New("not_implemented")
)
