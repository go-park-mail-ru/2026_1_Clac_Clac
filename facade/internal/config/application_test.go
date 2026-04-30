package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultApplicationConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		want := config.Application{
			LogLevel:           config.DebugLevel,
			MaxTextRequestSize: 10 * 1024,
			MaxUploadImageSize: 10 * 1024 * 1024,
		}

		actual := config.DefaultApplicationConfig()
		assert.Equal(t, want, actual)
	})
}

func TestApplicationConfigUnmarshal(t *testing.T) {
	t.Run("unmarshal from yaml", func(t *testing.T) {
		want := config.DefaultApplicationConfig()

		yaml := []byte(`
log_level: debug
max_text_request_size: 10240
max_upload_image_size: 10485760
`)

		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(yaml))
		require.NoError(t, err)

		conf := config.DefaultApplicationConfig()
		err = v.Unmarshal(&conf)
		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)
	})
}

func TestIsDebug(t *testing.T) {
	assert.True(t, config.IsDebug(config.DebugLevel), "debug level must return true")
	assert.False(t, config.IsDebug(config.InfoLevel), "info level must return false")
	assert.False(t, config.IsDebug("warn"), "unknown level must return false")
	assert.False(t, config.IsDebug(""), "empty level must return false")
}
