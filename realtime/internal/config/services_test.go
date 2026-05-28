package config_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServicesConfig(t *testing.T) {
	t.Run("all sub-configs present", func(t *testing.T) {
		want := config.Services{
			Auth:  config.DefaultAuthConfig(),
			Board: config.DefaultBoardConfig(),
		}

		actual := config.DefaultServicesConfig()
		assert.Equal(t, want, actual)
	})
}
