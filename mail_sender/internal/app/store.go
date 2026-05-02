package app

import (
	"fmt"

	sender "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/repository"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Sender *sender.Repository

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
		Sender: sender.NewRepository(redisClient),

		redisClient: redisClient,
	}
}
