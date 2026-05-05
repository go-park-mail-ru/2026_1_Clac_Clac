package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAuthConfig(t *testing.T) {
	t.Run("default handler values", func(t *testing.T) {
		conf := DefaultAuthConfig()

		assert.Equal(t, authConfigDefaultMaxLenPassword, conf.Handler.MaxLenPassword)
		assert.Equal(t, authConfigDefaultMinLenPassword, conf.Handler.MinLenPassword)
		assert.Equal(t, authConfigSessionLifeTime, conf.Handler.SessionLifetime)
		assert.Equal(t, authConfigDefaultVKOAuthRedirectTo, conf.Handler.VKOAuthRedirectTo)
	})

	t.Run("default client equals DefaultClientConfig", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})

	t.Run("session lifetime is 24h", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, 24*time.Hour, conf.Handler.SessionLifetime)
	})

	t.Run("password length bounds", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, 8, conf.Handler.MinLenPassword)
		assert.Equal(t, 128, conf.Handler.MaxLenPassword)
	})
}
