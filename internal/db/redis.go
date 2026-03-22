package db

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var ErrorConnectRadis = errors.New("cannot connect to Radis")

func NewPoolRedis(settings *redis.Options, logger *zerolog.Logger, timeBeforeRetry time.Duration) (*redis.Client, error) {
	const maxRetries = 5

	client := redis.NewClient(settings)

	for i := 1; i <= maxRetries; i++ {
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)

		err := client.Ping(contextWithTimeout).Err()
		cancel()

		if err == nil {
			logger.Info().Msgf("Successfully connected to Redis (Attempt %d)", i)
			return client, nil
		}

		logger.Warn().Msgf("Redis not ready yet, retrying")

		if i < maxRetries {
			time.Sleep(timeBeforeRetry)
		}
	}

	client.Close()
	return nil, ErrorConnectRadis
}
