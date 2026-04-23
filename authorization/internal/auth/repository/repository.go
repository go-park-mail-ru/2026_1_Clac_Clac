package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository/dto"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/common"
)

type RedisEngine interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

type Repository struct {
	redisClient RedisEngine
}

func NewRepository(redisClient RedisEngine) *Repository {
	return &Repository{
		redisClient: redisClient,
	}
}

func (r *Repository) AddSession(ctx context.Context, session dto.SessionEntity) error {
	err := r.redisClient.Set(ctx, session.SessionKey, session.UserLink.String(), session.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("redisClient.Set: %w", err)
	}

	return nil
}

func (r *Repository) GetUserLink(ctx context.Context, sessionKey string) (string, error) {
	userLink, err := r.redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingSession
		}

		return "", fmt.Errorf("redisClient.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionKey string) error {
	err := r.redisClient.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("redisClient.Del: %w", err)
	}

	return nil
}

func (r *Repository) ExtendSession(ctx context.Context, session dto.ExtendedSession) error {
	ok, err := r.redisClient.Expire(ctx, session.SessionKey, session.Expiration).Result()
	if err != nil {
		return fmt.Errorf("redisClient.Expire: %w", err)
	}

	if !ok {
		return common.ErrorNotExistingSession
	}

	return nil
}
