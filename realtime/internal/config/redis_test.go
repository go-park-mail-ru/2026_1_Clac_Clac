package config

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRedisConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := RedisConfig{
			NumberDB:       0,
			MinConnections: 20,
			MaxConnections: 100,
			PingSleepTime:  2 * time.Second,
			MaxRetries:     5,
		}

		actual := DefaultRedisConfig()
		assert.Equal(t, want, actual)
	})
}

func TestRedisConfigToPkg(t *testing.T) {
	t.Run("converts to pkg redis config", func(t *testing.T) {
		r := RedisConfig{
			Password:       "secret",
			Host:           "localhost",
			Port:           "6379",
			NumberDB:       0,
			MinConnections: 20,
			MaxConnections: 100,
			PingSleepTime:  2 * time.Second,
			MaxRetries:     5,
		}

		want := redis.Config{
			Password:       "secret",
			Host:           "localhost",
			Port:           "6379",
			NumberDB:       0,
			MinConnections: 20,
			MaxConnections: 100,
			PingSleepTime:  2 * time.Second,
			MaxRetries:     5,
		}

		assert.Equal(t, want, r.ToPkg())
	})
}
