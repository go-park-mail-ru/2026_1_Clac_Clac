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

	tests := []struct {
		nameTest      string
		options       *redis.Options
		expectedError error
	}{
		{
			nameTest: "Error connection timeout (exhaust retries)",
			options: &redis.Options{
				Addr:        "localhost:9999",
				DialTimeout: 100 * time.Millisecond,
			},
			expectedError: ErrorConnectRadis,
		},
	}

	for _, test := range tests {
		t.Run(test.nameTest, func(t *testing.T) {
			client, err := NewPoolRedis(test.options, &logger, timeBeforeRetry)

			assert.Nil(t, client, "client should be nil on error")
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}
