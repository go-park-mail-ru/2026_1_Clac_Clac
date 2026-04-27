package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSenderConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		want := Sender{
			Service: SenderService{
				CountRetries:       5,
				LifeTimeResetToken: 15 * time.Minute,
				SleepTime:          2 * time.Second,
			},
		}

		actual := DefaultSenderConfig()
		assert.Equal(t, want, actual)
	})
}
