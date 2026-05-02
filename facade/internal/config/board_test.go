package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBoardConfig(t *testing.T) {
	t.Run("handler defaults", func(t *testing.T) {
		conf := DefaultBoardConfig()
		assert.Equal(t, "", conf.Handler.MultipartBackgroundFileKey)
	})

	t.Run("client defaults", func(t *testing.T) {
		conf := DefaultBoardConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})
}
