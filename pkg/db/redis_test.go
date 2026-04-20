package db

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewPoolRedisError(t *testing.T) {
	logger := zerolog.Nop()

	pingSleepTime := time.Millisecond * 2
	maxRetries := 1

	tests := []struct {
		nameTest      string
		options       *redis.Options
		expectedError error
		errorContains string
	}{
		{
			nameTest: "Error connection timeout",
			options: &redis.Options{
				Addr:        "localhost:9999",
				DialTimeout: 10 * time.Millisecond,
			},
			expectedError: ErrorConnectRedis,
			errorContains: "",
		},
		{
			nameTest: "Error unknown host",
			options: &redis.Options{
				Addr:        "unknown-redis-host-123:6379",
				DialTimeout: 10 * time.Millisecond,
			},
			expectedError: ErrorConnectRedis,
			errorContains: "",
		},
		{
			nameTest: "Error invalid credentials",
			options: &redis.Options{
				Addr:        "localhost:6379",
				Password:    "wrong-super-secret-password",
				DialTimeout: 10 * time.Millisecond,
			},
			expectedError: ErrorConnectRedis,
			errorContains: "",
		},
		{
			nameTest: "Error empty address",
			options: &redis.Options{
				Addr:        "",
				DialTimeout: 10 * time.Millisecond,
			},
			expectedError: ErrorConnectRedis,
			errorContains: "",
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {

			client, err := NewPoolRedis(test.options, &RedisConnection{
				PingSleepTime: pingSleepTime,
				MaxRetries:    maxRetries,
			}, &logger)

			assert.Nil(t, client, "client should be nil on error")

			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
			}
			if test.errorContains != "" {
				assert.ErrorContains(t, err, test.errorContains)
			}
		})
	}
}
