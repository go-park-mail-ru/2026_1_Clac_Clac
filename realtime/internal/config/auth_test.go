package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAuthConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := Auth{
			SessionLifetime: 24 * time.Hour,
			Client: ClientAuth{
				ClientConfig: DefaultClientConfig(),
			},
		}

		actual := DefaultAuthConfig()
		assert.Equal(t, want, actual)
	})

	t.Run("session lifetime is 24h", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, 24*time.Hour, conf.SessionLifetime)
	})
}
