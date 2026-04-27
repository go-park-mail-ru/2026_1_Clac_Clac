package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		want := Auth{
			Service: AuthService{
				SessionLifetime: 24 * time.Hour,
			},
		}

		actual := DefaultAuthConfig()
		assert.Equal(t, want, actual)
	})
}
