package app

import (
	"fmt"

	limiter "github.com/go-park-mail-ru/2026_1_Clac_Clac/rate_limiter/internal/limiter/repository"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Limiter *limiter.Repository

	redisClient *redis.Client
}

func (s *Store) Close() error {
	err := s.redisClient.Close()
	if err != nil {
		return fmt.Errorf("cannot close redis client: %w", err)
	}

	return nil
}

func NewStore(redisClient *redis.Client) *Store {
	return &Store{
		Limiter: limiter.NewRepository(redisClient),

		redisClient: redisClient,
	}
}
