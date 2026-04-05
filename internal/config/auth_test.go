package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := config.Auth{
			CSRFSecret: "lsjdfklsdjf",
		}

		t.Setenv("AUTH_CSRF_SECRET", want.CSRFSecret)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvAuthConfig(v)

		var conf struct {
			Auth config.Auth `mapstructure:"auth"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.Auth)
	})
}
