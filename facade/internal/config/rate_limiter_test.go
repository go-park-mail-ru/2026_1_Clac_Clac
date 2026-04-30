package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultActionsRateLimiters(t *testing.T) {
	t.Run("default client config", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		assert.Equal(t, DefaultClientConfig(), conf.ClientConfig)
	})

	t.Run("login action exists", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		action, ok := conf.DBActions[LogInUser]
		require.True(t, ok, "login action must be present")
		assert.Equal(t, int64(defaultLimit), action.Limit)
		assert.Equal(t, LogInUser, action.Action)
		assert.Equal(t, 1*time.Minute, action.Window)
	})

	t.Run("register action exists", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		action, ok := conf.DBActions[RegisterUser]
		require.True(t, ok, "register action must be present")
		assert.Equal(t, int64(defaultLimit), action.Limit)
		assert.Equal(t, RegisterUser, action.Action)
		assert.Equal(t, 1*time.Hour, action.Window)
	})

}

func TestRateLimitersGetParameters(t *testing.T) {
	t.Run("returns login parameters", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		params := conf.GetParameters(LogInUser)

		assert.Equal(t, LogInUser, params.Action)
		assert.Equal(t, int64(defaultLimit), params.Limit)
		assert.Equal(t, 1*time.Minute, params.Window)
	})

	t.Run("returns register parameters", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		params := conf.GetParameters(RegisterUser)

		assert.Equal(t, RegisterUser, params.Action)
		assert.Equal(t, int64(defaultLimit), params.Limit)
		assert.Equal(t, 1*time.Hour, params.Window)
	})

	t.Run("returns zero value for unknown action", func(t *testing.T) {
		conf := DefaultActionsRateLimiters()
		params := conf.GetParameters("unknown_action")

		assert.Empty(t, params.Action)
		assert.Zero(t, params.Limit)
		assert.Zero(t, params.Window)
	})
}
