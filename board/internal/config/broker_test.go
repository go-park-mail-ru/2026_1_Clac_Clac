package config

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/stretchr/testify/assert"
)

func TestDefaultBrokerConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := BrokerConfig{
			NumberDB:       0,
			MinConnections: 20,
			MaxConnections: 100,
			PingSleepTime:  2 * time.Second,
			MaxRetries:     5,
		}

		actual := DefaultBrokerConfig()
		assert.Equal(t, want, actual)
	})
}

func TestBrokerConfigToPkg(t *testing.T) {
	t.Run("converts to pkg redis config", func(t *testing.T) {
		b := BrokerConfig{
			Password:       "broker-secret",
			Host:           "nexus-broker-service",
			Port:           "6379",
			NumberDB:       1,
			MinConnections: 10,
			MaxConnections: 50,
			PingSleepTime:  1 * time.Second,
			MaxRetries:     3,
		}

		want := redis.Config{
			Password:       "broker-secret",
			Host:           "nexus-broker-service",
			Port:           "6379",
			NumberDB:       1,
			MinConnections: 10,
			MaxConnections: 50,
			PingSleepTime:  1 * time.Second,
			MaxRetries:     3,
		}

		assert.Equal(t, want, b.ToPkg())
	})
}
