package db

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewPoolRedisError(t *testing.T) {
	logger := zerolog.Nop()

	pingSleepTime := time.Millisecond * 2
	maxRetries := 5

	tests := []struct {
		nameTest      string
		options       *redis.Options
		expectedError error
	}{
		{
			nameTest: "Error connection timeout",
			options: &redis.Options{
				Addr:        "localhost:9999",
				DialTimeout: 100 * time.Millisecond,
			},
			expectedError: ErrorConnectRedis,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			client, err := NewPoolRedis(test.options, &config.RedisConnection{PingSleepTime: pingSleepTime, MaxRetries: maxRetries},
				&logger)

			defer func() {
				errClose := client.Close()
				assert.NoError(t, errClose, "not wait error")
			}()

			assert.Nil(t, client, "client should be nil on error")
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}
