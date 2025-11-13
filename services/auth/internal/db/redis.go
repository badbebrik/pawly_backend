package db

import (
	"auth/internal/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	addr := fmt.Sprintf("%s:%s", cfg.PostgresHost, cfg.RedisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Info().Str("host", cfg.RedisHost).Str("db", strconv.Itoa(cfg.RedisDB)).Msg("Connected to redis")

	return &Redis{client: rdb}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}
