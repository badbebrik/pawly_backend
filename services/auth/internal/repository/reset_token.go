package repository

import (
	"context"
	"time"
)

type ResetTokenRepository interface {
	ConsumeOnce(ctx context.Context, tokenID string, ttl time.Duration) (bool, error)
}
