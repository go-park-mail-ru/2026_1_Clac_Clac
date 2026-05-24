package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBoardConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := Board{
			Client: ClientBoard{
				ClientConfig: DefaultClientConfig(),
			},
		}

		actual := DefaultBoardConfig()
		assert.Equal(t, want, actual)
	})
}
