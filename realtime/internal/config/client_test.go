package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClientConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := ClientConfig{
			Addr:    "",
			TimeOut: defaultTimeOut,
			Retries: defaultRetries,
		}

		actual := DefaultClientConfig()
		assert.Equal(t, want, actual)
	})

	t.Run("default timeout is 5s", func(t *testing.T) {
		conf := DefaultClientConfig()
		assert.Equal(t, 5*time.Second, conf.TimeOut)
	})

	t.Run("default retries is 5", func(t *testing.T) {
		conf := DefaultClientConfig()
		assert.Equal(t, 5, conf.Retries)
	})
}
