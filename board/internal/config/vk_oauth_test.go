package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVkOAuthConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.VkOAuth{
			AppID:       "1234",
			AppKey:      "jsldjf",
			AppSecret:   "lsdlksjdfkj",
			RedirectURL: "hsldhlk",
			APIMethod:   "lsdslkdjf",
		}

		t.Setenv("VK_OAUTH_APP_ID", want.AppID)
		t.Setenv("VK_OAUTH_APP_KEY", want.AppKey)
		t.Setenv("VK_OAUTH_APP_SECRET", want.AppSecret)
		t.Setenv("VK_OAUTH_REDIRECT_URL", want.RedirectURL)
		t.Setenv("VK_OAUTH_API_METHOD", want.APIMethod)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvVkOAuth(v)

		var conf struct {
			VkOAuth config.VkOAuth `mapstructure:"vk_oauth"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.VkOAuth)
	})
}
