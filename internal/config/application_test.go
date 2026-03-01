package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationConfig(t *testing.T) {
	t.Run("test unmarshal", func(t *testing.T) {
		want := config.Application{
			LogLevel: DebugLevel,
		}
		var yamlTest = []byte(`log_level: debug`)

		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yamlTest))

		conf := config.DefaultApplicationConfig()
		err := viper.Unmarshal(&conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)

	})
}

func TestIsDebug(t *testing.T) {
	assert.True(t, config.IsDebug(DebugLevel), "this is debug level")
	assert.False(t, config.IsDebug(InfoLevel), "this is not debug level")
}
