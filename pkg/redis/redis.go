<<<<<<< HEAD
package redis
=======
package db
>>>>>>> feat/add-facade

import (
	"context"
	"errors"
	"fmt"
	"time"

<<<<<<< HEAD
<<<<<<<< HEAD:pkg/redis/redis.go
========
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/monolith/internal/config"
>>>>>>>> feat/add-facade:monolith/internal/db/redis.go
=======
>>>>>>> feat/add-facade
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var ErrorConnectRedis = errors.New("cannot connect to Redis")

<<<<<<< HEAD
func NewPoolRedis(settings *redis.Options, conf *Config, logger *zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(settings)

	for i := 1; i <= conf.MaxRetries; i++ {
=======
type Config struct {
	PingSleepTime time.Duration
	MaxRetries    int
}

func NewPoolRedis(settings *redis.Options, redisConnection Config, logger *zerolog.Logger) (*redis.Client, error) {
	client := redis.NewClient(settings)

	for i := 1; i <= redisConnection.MaxRetries; i++ {
>>>>>>> feat/add-facade
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)

		err := client.Ping(contextWithTimeout).Err()
		cancel()

		if err == nil {
			logger.Info().Msgf("Successfully connected to Redis (Attempt %d)", i)
			return client, nil
		}

		logger.Warn().Msgf("Redis not ready yet, retrying")

<<<<<<< HEAD
		if i < conf.MaxRetries {
			time.Sleep(conf.PingSleepTime)
=======
		if i < redisConnection.MaxRetries {
			time.Sleep(redisConnection.PingSleepTime)
>>>>>>> feat/add-facade
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
