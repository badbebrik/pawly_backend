package ports

import "errors"

var (
	ErrNotFound = errors.New("not_found")
	ErrConflict = errors.New("conflict")
)
