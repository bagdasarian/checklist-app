package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bagdasarian/checklist-app/db_service/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
	TTL    time.Duration
}

func NewRedis(ctx context.Context, cfg *config.Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Проверяем подключение
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	ttl := time.Duration(cfg.Redis.TTL) * time.Second
	if ttl == 0 {
		ttl = 5 * time.Minute // По умолчанию 5 минут
	}

	log.Printf("Successfully connected to Redis at %s", cfg.GetRedisAddr())
	return &Redis{
		Client: rdb,
		TTL:    ttl,
	}, nil
}

func (r *Redis) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
