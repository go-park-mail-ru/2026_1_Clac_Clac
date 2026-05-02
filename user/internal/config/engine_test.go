package config_test

import (
	"bytes"
	"testing"

	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineConfig(t *testing.T) {
	t.Run("test unmarshal", func(t *testing.T) {
		want := engine.Config{
			Addr:                    ":50053",
			GracefulShutdownTimeout: 15,
		}

		var yamlTest = []byte(`
addr: ":50053"
graceful_shutdown_timeout: 15
`)

		viper.SetConfigType("yaml")
		viper.ReadConfig(bytes.NewBuffer(yamlTest))

		conf := config.DefaultEngineConfig()
		err := viper.Unmarshal(&conf)

		require.NoError(t, err, "viper.Unmarshal should not return error")
		assert.Equal(t, want, conf)
	})
}
