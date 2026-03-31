package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultActionsRateLimiters(t *testing.T) {
	t.Run("Default init", func(t *testing.T) {
		expectedDB := DataBaseRateLimiters{
			DBActions: map[string]ActionRateLimiters{
				LogInUser: {
					Limit:  defaultLimit,
					Action: LogInUser,
					Window: 1 * time.Minute,
				},
				RegisterUser: {
					Limit:  defaultLimit,
					Action: RegisterUser,
					Window: 1 * time.Hour,
				},
			},
		}

		db := DefaultActionsRateLimiters()

		assert.Equal(t, expectedDB, db)
	})
}
