package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type Repository interface {
	RequestCode(ctx context.Context, email, purpose string) (code string, ttlSeconds int, resendInSeconds int, err error)
	VerifyCode(ctx context.Context, email, purpose, code string) error
}

const (
	CodeTTL        = 15 * time.Minute
	ResendCooldown = 60 * time.Second
	MaxAttempts    = 5
)

type RedisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{rdb: rdb}
}

func key(purpose, email string) string {
	return fmt.Sprintf("verification:%s:%s", purpose, email)
}

func (r *RedisRepository) RequestCode(ctx context.Context, email, purpose string) (string, int, int, error) {
	k := key(purpose, email)

	now := time.Now()

	data, err := r.rdb.Get(ctx, k).Bytes()
	if err == nil {
		var rec CodeRecord
		if err := json.Unmarshal(data, &rec); err == nil {
			if now.Before(rec.ResendAvailableAt) {
				return "", int(CodeTTL.Seconds()), int(rec.ResendAvailableAt.Sub(now).Seconds()), ErrResendTooSoon
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		return "", 0, 0, err
	}

	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	rec := CodeRecord{
		Code:              code,
		CreatedAt:         now,
		ExpiresAt:         now.Add(CodeTTL),
		ResendAvailableAt: now.Add(ResendCooldown),
		Attempts:          0,
	}

	jsonData, err := json.Marshal(rec)
	if err != nil {
		return "", 0, 0, err
	}

	if err := r.rdb.Set(ctx, k, jsonData, CodeTTL).Err(); err != nil {
		return "", 0, 0, err
	}

	return code, int(CodeTTL.Seconds()), int(ResendCooldown.Seconds()), nil
}

func (r *RedisRepository) VerifyCode(ctx context.Context, email, purpose, inputCode string) error {
	k := key(purpose, email)
	now := time.Now()

	data, err := r.rdb.Get(ctx, k).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCodeNotFound
		}
		return err
	}

	var rec CodeRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return err
	}

	if now.After(rec.ExpiresAt) {
		r.rdb.Del(ctx, k)
		return ErrCodeExpired
	}

	if rec.Attempts >= MaxAttempts {
		r.rdb.Del(ctx, k)
		return ErrTooManyAttempts
	}

	if inputCode != rec.Code {
		rec.Attempts++

		jsonData, _ := json.Marshal(rec)
		r.rdb.Set(ctx, k, jsonData, rec.ExpiresAt.Sub(now))

		return ErrCodeInvalid
	}

	r.rdb.Del(ctx, k)

	return nil
}
