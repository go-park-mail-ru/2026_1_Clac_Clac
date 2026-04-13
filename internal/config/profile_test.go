package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfileConfig(t *testing.T) {

	t.Run("Default config", func(t *testing.T) {
		want := Profile{
			Handler: ProfileHandler{
				SiganatureTypeBytes:   512,
				MaxReadBytes:          5242880,
				MaxLenNameUser:        128,
				MaxLenDescriptionUser: 500,
			},
		}

		actual := DefaultProfileConfig()
		assert.Equal(t, want, actual)
	})
}
