package verification

import "time"

const (
	CodeTTL        = 15 * time.Minute
	ResendCooldown = 60 * time.Second
	MaxAttempts    = 5
)
