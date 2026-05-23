package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCardConfig(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		want := Card{
			Handler: CardHandler{
				MaxLenTitle:       128,
				MaxLenDescription: 500,
			},
			Repository: CardRepository{
				MaxAttachments:  100,
				MaxNestingDepth: 100,
			},
		}

		actual := DefaultCardConfig()
		assert.Equal(t, want, actual)
	})
}
