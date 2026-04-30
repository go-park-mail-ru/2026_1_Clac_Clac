package config_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServicesConfig(t *testing.T) {
	t.Run("all sub-configs are initialized", func(t *testing.T) {
		conf := config.DefaultServicesConfig()

		assert.Equal(t, config.DefaultClientConfig(), conf.MailSender)
		assert.Equal(t, config.DefaultUserConfig(), conf.User)
		assert.Equal(t, config.DefaultAuthConfig(), conf.Auth)
		assert.Equal(t, config.DefaultClientConfig(), conf.Board)
		assert.Equal(t, config.DefaultActionsRateLimiters(), conf.RateLimiters)
	})
}
