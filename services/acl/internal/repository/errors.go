package repository

import "errors"

var (
	ErrNotFound  = errors.New("not_found")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
)
