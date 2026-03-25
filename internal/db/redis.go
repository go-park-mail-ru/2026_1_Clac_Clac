package db

import (
	"context"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var ErrorConnectRadis = errors.New("cannot connect to Redis")

func NewPoolRedis(settings *redis.Options, redisConnection *config.RedisConnection, logger *zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(settings)

	for i := 1; i <= redisConnection.MaxRetries; i++ {
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)

		err := client.Ping(contextWithTimeout).Err()
		cancel()

		if err == nil {
			logger.Info().Msgf("Successfully connected to Redis (Attempt %d)", i)
			return client, nil
		}

		logger.Warn().Msgf("Redis not ready yet, retrying")

		if i < redisConnection.MaxRetries {
			time.Sleep(redisConnection.PingSleepTime)
		}
	}

	client.Close()
	return nil, ErrorConnectRadis
}
