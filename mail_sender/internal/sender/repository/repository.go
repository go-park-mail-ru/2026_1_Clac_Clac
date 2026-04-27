package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/common"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository/dto"
	"github.com/redis/go-redis/v9"
)

type RedisEngine interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type Repository struct {
	redisClient RedisEngine
}

func NewRepository(client RedisEngine) *Repository {
	return &Repository{
		redisClient: client,
	}
}

func (r *Repository) AddResetToken(ctx context.Context, token dto.ResetTokenEntity) error {
	err := r.redisClient.Set(ctx, token.ResetTokenKey, token.UserLink.String(), token.LifeTime).Err()
	if err != nil {
		return fmt.Errorf("client.Set: %w", err)
	}

	return nil
}

func (r *Repository) GetUserLinkByResetToken(ctx context.Context, tokenKey string) (string, error) {
	userLink, err := r.redisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", common.ErrorNotExistingResetToken
		}

		return "", fmt.Errorf("client.Get: %w", err)
	}

	return userLink, nil
}

func (r *Repository) DeleteResetToken(ctx context.Context, tokenKey string) error {
	err := r.redisClient.Del(ctx, tokenKey).Err()
	if err != nil {
		return fmt.Errorf("client.Del: %w", err)
	}

	return nil
}
