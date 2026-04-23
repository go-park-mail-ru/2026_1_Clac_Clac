package app

import (
	"fmt"

	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Auth *auth.Repository

	redisClient *redis.Client
}

func (s *Store) Close() error {
	err := s.redisClient.Close()
	if err != nil {
		return fmt.Errorf("cannot close redis client: %w", err)
	}

	return nil
}

func NewStore(redisClient *redis.Client, conf config.Config) *Store {
	return &Store{
		Auth: auth.NewRepository(redisClient),

		redisClient: redisClient,
	}
}
