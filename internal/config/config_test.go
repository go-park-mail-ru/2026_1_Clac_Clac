package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWithViper(t *testing.T) {
	want := &config.ApplicationConfig{
		Debug: false,
	}
	var yamlTest = []byte(`
app:
  debug: false
`)

	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(yamlTest))

	require.NoError(t, err, "reading should not returt error")

	conf := config.DefaultApplicationConfig()
	err = config.ReadWithViper(v, conf)

	require.NoError(t, err, "should not return error")
	assert.Equal(t, want, conf)
}

func TestMultipleConfigReading(t *testing.T) {
	expectedAppConfig := &config.ApplicationConfig{
		Debug: false,
	}
	expectedEngineConfig := &config.EngineConfig{
		Addr:                    ":8080",
		WriteTimeout:            30,
		ReadTimeout:             30,
		IdleTimeout:             90,
		GracefulShutdownTimeout: 25,
	}
	var yamlTest = []byte(`
app:
  debug: false

http:
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

	appConf := config.DefaultApplicationConfig()
	err = config.ReadWithViper(v, appConf)

	require.NoError(t, err, "error while setting app config")
	assert.Equal(t, expectedAppConfig, appConf)

	engineConf := config.DefaultEngineConfig()
	err = config.ReadWithViper(v, engineConf)

	require.NoError(t, err, "error while setting engine config")
	assert.Equal(t, expectedEngineConfig, engineConf)
}
