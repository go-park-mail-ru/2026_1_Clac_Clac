package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAuthConfig(t *testing.T) {
	t.Run("default handler values", func(t *testing.T) {
		conf := DefaultAuthConfig()

		assert.Equal(t, authConfigDefaultMaxLenPassword, conf.Handler.MaxLenPassword)
		assert.Equal(t, authConfigDefaultMinLenPassword, conf.Handler.MinLenPassword)
		assert.Equal(t, authConfigSessionLifeTime, conf.Handler.SessionLifetime)
		assert.Equal(t, authConfigDefaultVKOAuthRedirectTo, conf.Handler.VKOAuthRedirectTo)
	})

	t.Run("default client equals DefaultClientConfig", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, DefaultClientConfig(), conf.Client.ClientConfig)
	})

	t.Run("session lifetime is 24h", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, 24*time.Hour, conf.Handler.SessionLifetime)
	})

	t.Run("password length bounds", func(t *testing.T) {
		conf := DefaultAuthConfig()
		assert.Equal(t, 8, conf.Handler.MinLenPassword)
		assert.Equal(t, 128, conf.Handler.MaxLenPassword)
	})
}

func TestSetupEnvAuth(t *testing.T) {
	t.Run("vk oauth redirect from env", func(t *testing.T) {
		redirectURL := "https://example.com/oauth/callback"
		t.Setenv("SERVICES_AUTH_HANDLER_VK_OAUTH_REDIRECT_TO", redirectURL)

		v := viper.New()

		SetupEnvAuth(v)

		var conf struct {
			Services struct {
				Auth Auth `mapstructure:"auth"`
			} `mapstructure:"services"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper.Unmarshal must not return error")
		assert.Equal(t, redirectURL, conf.Services.Auth.Handler.VKOAuthRedirectTo)
	})
}
