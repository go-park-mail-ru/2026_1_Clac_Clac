package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/config"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSentryConfig(t *testing.T) {
	t.Run("returns correct default values", func(t *testing.T) {
		expected := sentryLogger.Sentry{
			Environment: "production",
			Release:     "RK3",
			ServiceName: "appeal",
			Tags: map[string]string{
				"layer":    "grpc",
				"team":     "api-platform",
				"protocol": "grpc",
			},
			TracesSampleRate: 0.1,
			Repanic:          true,
		}

		actual := config.DefaultSentryConfig()
		assert.Equal(t, expected, actual)
	})

	t.Run("DSN is empty by default", func(t *testing.T) {
		cfg := config.DefaultSentryConfig()
		assert.Empty(t, cfg.DSN)
	})

	t.Run("environment is production", func(t *testing.T) {
		cfg := config.DefaultSentryConfig()
		assert.Equal(t, "production", cfg.Environment)
	})

	t.Run("service name is appeal", func(t *testing.T) {
		cfg := config.DefaultSentryConfig()
		assert.Equal(t, "appeal", cfg.ServiceName)
	})

	t.Run("repanic is enabled", func(t *testing.T) {
		cfg := config.DefaultSentryConfig()
		assert.True(t, cfg.Repanic)
	})
}

func TestSetupEnvSentryConfig(t *testing.T) {
	t.Run("default DSN is empty string", func(t *testing.T) {
		v := viper.New()
		config.SetupEnvSentryConfig(v)

		assert.Equal(t, "", v.GetString("sentry.dsn"))
	})

	t.Run("SENTRY_DSN env var is read via alias", func(t *testing.T) {
		const testDSN = "https://key@sentry.example.io/123"
		t.Setenv("SENTRY_DSN", testDSN)

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
		config.SetupEnvSentryConfig(v)

		var conf struct {
			Sentry sentryLogger.Sentry `mapstructure:"sentry"`
		}
		err := v.Unmarshal(&conf)
		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, testDSN, conf.Sentry.DSN)
	})

	t.Run("alias sentry_dsn maps to sentry.dsn", func(t *testing.T) {
		const overrideDSN = "https://override@sentry.example.io/456"
		v := viper.New()
		config.SetupEnvSentryConfig(v)
		v.Set("sentry_dsn", overrideDSN)

		assert.Equal(t, overrideDSN, v.GetString("sentry.dsn"))
	})
}
