package redisstore

import (
	"auth/internal/application/ports"
	"auth/internal/verification"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/big"
	"time"
)

type Store struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func key(purpose, email string) string {
	return fmt.Sprintf("verification:%s:%s", purpose, email)
}

func (s *Store) RequestCode(ctx context.Context, email, purpose string) (string, int, int, error) {
	k := key(purpose, email)

	now := time.Now()

	data, err := s.rdb.Get(ctx, k).Bytes()
	if err == nil {
		var rec verification.CodeRecord
		if err := json.Unmarshal(data, &rec); err == nil {
			if now.Before(rec.ResendAvailableAt) {
				return "", int(verification.CodeTTL.Seconds()), int(rec.ResendAvailableAt.Sub(now).Seconds()), ports.ErrResendTooSoon
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		return "", 0, 0, err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", 0, 0, err
	}
	code := fmt.Sprintf("%06d", n.Int64())

	rec := verification.CodeRecord{
		Code:              code,
		CreatedAt:         now,
		ExpiresAt:         now.Add(verification.CodeTTL),
		ResendAvailableAt: now.Add(verification.ResendCooldown),
		Attempts:          0,
	}

	jsonData, err := json.Marshal(rec)
	if err != nil {
		return "", 0, 0, err
	}

	if err := s.rdb.Set(ctx, k, jsonData, verification.CodeTTL).Err(); err != nil {
		return "", 0, 0, err
	}

	return code, int(verification.CodeTTL.Seconds()), int(verification.ResendCooldown.Seconds()), nil
}

func (s *Store) VerifyCode(ctx context.Context, email, purpose, inputCode string) error {
	k := key(purpose, email)
	now := time.Now()

	data, err := s.rdb.Get(ctx, k).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ports.ErrCodeNotFound
		}
		return err
	}

	var rec verification.CodeRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return err
	}

	if now.After(rec.ExpiresAt) {
		s.rdb.Del(ctx, k)
		return ports.ErrCodeExpired
	}

	if rec.Attempts >= verification.MaxAttempts {
		s.rdb.Del(ctx, k)
		return ports.ErrTooManyAttempts
	}

	if subtle.ConstantTimeCompare([]byte(inputCode), []byte(rec.Code)) != 1 {
		rec.Attempts++

		jsonData, _ := json.Marshal(rec)
		s.rdb.Set(ctx, k, jsonData, rec.ExpiresAt.Sub(now))

		return ports.ErrCodeInvalid
	}

	s.rdb.Del(ctx, k)

	return nil
}
