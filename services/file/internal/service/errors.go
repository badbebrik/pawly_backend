package service

import (
	"errors"
)

var (
	ErrInvalidInput  = errors.New("invalid_input")
	ErrInvalidState  = errors.New("invalid_state")
	ErrUploadExpired = errors.New("upload_expired")
	ErrNotReady      = errors.New("not_ready")
)
