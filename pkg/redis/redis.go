package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var ErrorConnectRedis = errors.New("cannot connect to Redis")

func NewPoolRedis(settings *redis.Options, conf *Config, logger *zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(settings)

	for i := 1; i <= conf.MaxRetries; i++ {
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)

		err := client.Ping(contextWithTimeout).Err()
		cancel()

		if err == nil {
			logger.Info().Msgf("Successfully connected to Redis (Attempt %d)", i)
			return client, nil
		}

		logger.Warn().Msgf("Redis not ready yet, retrying")

		if i < conf.MaxRetries {
			time.Sleep(conf.PingSleepTime)
		}
	}

	err := ErrorConnectRedis

	errClose := client.Close()
	if errClose != nil {
		errClose = fmt.Errorf("cannot close client: %w", errClose)
		err = errors.Join(err, errClose)
	}

	return nil, err
}
