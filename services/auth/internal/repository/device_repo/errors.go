package device_repo

import "errors"

var (
	ErrDeviceNotFound = errors.New("device not found")
	ErrTokenInUse     = errors.New("fcm token is already used by another user")
)
