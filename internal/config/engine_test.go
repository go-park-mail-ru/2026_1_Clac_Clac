package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineConfig(t *testing.T) {
	t.Run("test section", func(t *testing.T) {
		conf := config.DefaultEngineConfig()
		assert.Equal(t, "http", conf.Section())
	})

	t.Run("test unmarshal", func(t *testing.T) {
		want := &config.EngineConfig{
			Addr:                    ":8080",
			WriteTimeout:            30,
			ReadTimeout:             30,
			IdleTimeout:             90,
			GracefulShutdownTimeout: 25,
		}
		var yamlTest = []byte(`
http:
  addr: ":8080"
  write_timeout: 30
  read_timeout: 30
  idle_timeout: 90
  graceful_shutdown_timeout: 25
`)

		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yamlTest))

		conf := config.DefaultEngineConfig()
		httpSection := viper.Sub(conf.Section())

		err := httpSection.Unmarshal(conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)
	})
}
