package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConnectionConfig(t *testing.T) {
	t.Run("test reading redis config from env", func(t *testing.T) {
		expected := RedisConnection{
			Password: "X5NiUyrTxlWmwK8BFpYq",
			Host:     "localhost",
			Port:     "6379",
		}

		t.Setenv("REDIS_PASSWORD", expected.Password)
		t.Setenv("REDIS_HOST", expected.Host)
		t.Setenv("REDIS_PORT", expected.Port)

		v := viper.New()

		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		SetupEnvRedisConnection(v)

		var conf struct {
			Database RedisConnection `mapstructure:"redis"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, expected, conf.Database, "expected other comfiguration redis")
	})
}
