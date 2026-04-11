package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSectionConfig(t *testing.T) {

	t.Run("Default config", func(t *testing.T) {
		want := Section{
			Handler: SectionHandler{
				MaxQuantityTasks:  100,
				MinQuantityTasks:  0,
				MaxLenNameSection: 128,
			},
		}

		actual := DefaultSectionConfig()
		assert.Equal(t, want, actual)
	})
}
