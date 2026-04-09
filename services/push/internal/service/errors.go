package service

import "errors"

var (
	ErrForbidden  = errors.New("forbidden")
	ErrNotFound   = errors.New("not_found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation")
)
