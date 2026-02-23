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
	t.Run("test section", func(t *testing.T) {
		conf := config.DefaultApplicationConfig()
		assert.Equal(t, "app", conf.Section())
	})

	t.Run("test unmarshal", func(t *testing.T) {
		want := &config.ApplicationConfig{
			Debug: false,
		}
		var yamlTest = []byte(`
app:
  debug: false
`)

		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yamlTest))

		conf := config.DefaultApplicationConfig()
		appSection := viper.Sub(conf.Section())

		err := appSection.Unmarshal(conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)

	})
}
