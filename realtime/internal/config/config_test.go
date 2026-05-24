package config_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("all sub-configs present", func(t *testing.T) {
		conf := config.DefaultConfig()

		assert.Equal(t, config.DefaultApplicationConfig(), conf.App)
		assert.Equal(t, config.DefaultRedisConfig(), conf.Redis)
		assert.Equal(t, config.DefaultServicesConfig(), conf.Services)
	})
}

func TestSetupViper(t *testing.T) {
	const configFilename = "config.yaml"

	minimalYAML := []byte(`
app:
  log_level: info
  max_message_size: 4194304
`)

	t.Run("reads config from directory", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, configFilename), minimalYAML, 0644)
		require.NoError(t, err)

		v, err := config.SetupViper(tempDir)
		require.NoError(t, err, "SetupViper must not return error when config.yaml exists")
		require.NotNil(t, v)
	})

	t.Run("unmarshal into default config", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, configFilename), minimalYAML, 0644)
		require.NoError(t, err)

		v, err := config.SetupViper(tempDir)
		require.NoError(t, err)

		conf := config.DefaultConfig()
		err = v.Unmarshal(&conf)
		require.NoError(t, err, "Unmarshal must not return error")
		assert.Equal(t, config.InfoLevel, conf.App.LogLevel)
	})

	t.Run("returns error when config.yaml is missing", func(t *testing.T) {
		emptyDir := t.TempDir()
		_, err := config.SetupViper(emptyDir)
		require.Error(t, err, "SetupViper must return error when config.yaml is absent")
	})
}

func TestConfigReadingWithViper(t *testing.T) {
	t.Run("partial yaml keeps default values", func(t *testing.T) {
		defaultConf := config.DefaultConfig()

		partialYAML := []byte(`
app:
  log_level: info
`)

		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(partialYAML))
		require.NoError(t, err)

		conf := config.DefaultConfig()
		err = v.Unmarshal(&conf)
		require.NoError(t, err)

		assert.Equal(t, "info", conf.App.LogLevel)
		assert.Equal(t, defaultConf.App.MaxMessageSize, conf.App.MaxMessageSize)
	})
}
