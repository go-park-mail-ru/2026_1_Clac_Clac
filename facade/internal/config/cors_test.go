package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCORSConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := CORS{
			Credentials: defaultCorsCredentials,
			Origin:      defaultCorsOrigin,
			Methods:     defaultCorsMethods,
			Headers:     defaultCorsHeaders,
			MaxAge:      defaultCorsMaxAge,
		}

		actual := DefaultCORSConfig()
		assert.Equal(t, want, actual)
	})
}

func TestSetupEnvCORS(t *testing.T) {
	t.Run("origin from env", func(t *testing.T) {
		expectedOrigin := "https://example.com"
		t.Setenv("CORS_ORIGIN", expectedOrigin)

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		SetupEnvCORS(v)

		var conf struct {
			CORS CORS `mapstructure:"cors"`
		}

		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper.Unmarshal must not return error")
		assert.Equal(t, expectedOrigin, conf.CORS.Origin)
	})

	t.Run("default CORS config contains correct origin", func(t *testing.T) {
		conf := DefaultCORSConfig()
		assert.Equal(t, defaultCorsOrigin, conf.Origin)
	})
}
