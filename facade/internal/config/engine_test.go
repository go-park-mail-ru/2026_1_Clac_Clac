package config_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultEngineConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		conf := config.DefaultEngineConfig()

		assert.Equal(t, "localhost:8081", conf.Addr)
		assert.Equal(t, 15*time.Second, conf.WriteTimeout)
		assert.Equal(t, 15*time.Second, conf.ReadTimeout)
		assert.Equal(t, 60*time.Second, conf.IdleTimeout)
		assert.Equal(t, 15*time.Second, conf.GracefulShutdownTimeout)
	})
}

func TestEngineConfigUnmarshal(t *testing.T) {
	t.Run("unmarshal from yaml", func(t *testing.T) {
		yaml := []byte(`
addr: "0.0.0.0:8081"
write_timeout: 30s
read_timeout: 30s
idle_timeout: 120s
graceful_shutdown_timeout: 10s
`)

		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(bytes.NewBuffer(yaml))
		require.NoError(t, err)

		conf := config.DefaultEngineConfig()
		err = v.Unmarshal(&conf)
		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, "0.0.0.0:8081", conf.Addr)
		assert.Equal(t, 30*time.Second, conf.WriteTimeout)
	})
}
