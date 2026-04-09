package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthConfig(t *testing.T) {
	t.Run("test reading from env", func(t *testing.T) {
		want := Auth{
			Handler: AuthHandler{
				MaxLenPassword:  0,
				MinLenPassword:  0,
				SessionLifetime: 0,
			},
			Service: AuthService{
				CSRFSecret:      "lsjdfklsdjf",
				SessionLifetime: 0,
				CountRetries:    0,
			},
		}

		t.Setenv("AUTH_SERVICE_CSRF_SECRET", want.Service.CSRFSecret)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		SetupEnvAuthConfig(v)

		var conf struct {
			Auth Auth `mapstructure:"auth"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, want, conf.Auth)
	})
}
