package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

func TestConfigReading(t *testing.T) {
	expectedConfig := config.Config{
		App: config.Application{
			LogLevel: DebugLevel,
		},
		Engine: config.Engine{
			Addr:                    ":8080",
			WriteTimeout:            30,
			ReadTimeout:             30,
			IdleTimeout:             90,
			GracefulShutdownTimeout: 25,
		},
	}

	var yamlTest = []byte(`
app:
  log_level: debug

engine:
  addr: ":8080"
  write_timeout: 30
  read_timeout: 30
  idle_timeout: 90
  graceful_shutdown_timeout: 25
`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yamlTest))

	require.NoError(t, err, "reading should not returt error")

	conf := config.DefaultConfig()
	err = v.Unmarshal(&conf)

	require.NoError(t, err, "error while reading config")
	assert.Equal(t, expectedConfig, conf)
}

func TestSetupViper(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		const configFilename = "config.yaml"

		var yamlTest = []byte(`
app:
  log_level: debug

engine:
  addr: ":8080"
  write_timeout: 30
  read_timeout: 30
  idle_timeout: 90
  graceful_shutdown_timeout: 25
`)
		tempDir := t.TempDir()
		os.WriteFile(filepath.Join(tempDir, configFilename), yamlTest, 0644)

		v, err := config.SetupViper(tempDir)

		require.NoError(t, err, "must not return error")

		conf := config.DefaultConfig()
		err = v.Unmarshal(&conf)

		require.NoError(t, err, "error while reading config")
	})

	t.Run("error", func(t *testing.T) {
		const currentDir = "."
		_, err := config.SetupViper(currentDir)

		require.Error(t, err, "must return error")
	})
}
