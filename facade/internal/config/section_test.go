package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSectionConfig(t *testing.T) {
	t.Run("client defaults", func(t *testing.T) {
		conf := DefaultSectionConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})
}