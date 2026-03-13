package redisdb

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type ResetTokenStore struct {
	client *redis.Client
}

func NewResetTokenStore(client *redis.Client) *ResetTokenStore {
	return &ResetTokenStore{client: client}
}

func (s *ResetTokenStore) ConsumeOnce(ctx context.Context, tokenID string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = time.Minute
	}

	ok, err := s.client.SetNX(ctx, "password_reset_token:used:"+tokenID, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}
