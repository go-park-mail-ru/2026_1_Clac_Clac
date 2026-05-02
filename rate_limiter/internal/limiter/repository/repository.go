package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository/dto"
	"github.com/redis/go-redis/v9"
)

type RedisEngine interface {
	Pipeline() redis.Pipeliner
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
}

type Repository struct {
	redisClient RedisEngine
}

func NewRepository(client RedisEngine) *Repository {
	return &Repository{
		redisClient: client,
	}
}

func (r *Repository) CheckLimit(ctx context.Context, config dto.RateLimiterConfig) (int64, error) {
	now := time.Now().UnixNano()
	bucket := now / config.Window.Nanoseconds()
	key := fmt.Sprintf("rl:%s:%s:%d", config.Action, config.UserIP, bucket)

	pipe := r.redisClient.Pipeline()
	counter := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, config.Window)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("pipe.Exec: %w", err)
	}

	return counter.Val(), nil
}

func (r *Repository) SetCooldown(ctx context.Context, config dto.CooldownConfig) (bool, time.Duration, error) {
	isSet, err := r.redisClient.SetNX(ctx, config.Key, "", config.Expiration).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redisClient.SetNX: %w", err)
	}

	if isSet {
		return true, 0, nil
	}

	ttl, err := r.redisClient.TTL(ctx, config.Key).Result()
	if err != nil {
		return false, 0, fmt.Errorf("redisClient.TTL: %w", err)
	}

	if ttl < 0 {
		ttl = 0
	}

	return false, ttl, nil
}
