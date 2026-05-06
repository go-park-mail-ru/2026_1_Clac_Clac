package config_test

import (
	"bytes"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMetricsConfig(t *testing.T) {
	t.Run("default port", func(t *testing.T) {
		conf := config.DefaultMetrics()
		assert.Equal(t, ":9091", conf.MetricsPort)
	})
}

func TestMetricsConfigUnmarshal(t *testing.T) {
	t.Run("unmarshal from yaml", func(t *testing.T) {
		want := config.Metrics{MetricsPort: ":9999"}

		yaml := []byte(`metrics_port: ":9999"`)

		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(yaml))
		require.NoError(t, err)

		conf := config.DefaultMetrics()
		err = v.Unmarshal(&conf)
		require.NoError(t, err)
		assert.Equal(t, want, conf)
	})

	t.Run("unmarshal default values", func(t *testing.T) {
		want := config.DefaultMetrics()

		yaml := []byte(`metrics_port: ":9091"`)

		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(yaml))
		require.NoError(t, err)

		conf := config.DefaultMetrics()
		err = v.Unmarshal(&conf)
		require.NoError(t, err)
		assert.Equal(t, want, conf)
	})
}
