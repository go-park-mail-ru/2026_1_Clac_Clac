package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		want := User{
			Handler: UserHandler{
				MaxLenPassword:        128,
				MinLenPassword:        8,
				SiganatureTypeBytes:   512,
				MaxReadBytes:          5 << 20,
				MaxLenNameUser:        128,
				MaxLenDescriptionUser: 500,
			},
		}

		actual := DefaultUserConfig()
		assert.Equal(t, want, actual)
	})
}
