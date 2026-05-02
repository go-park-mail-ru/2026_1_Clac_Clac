package config

import (
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConnectionConfig(t *testing.T) {
	t.Run("test reading redis config from env", func(t *testing.T) {
		expected := redis.Config{
			Password: "X5NiUyrTxlWmwK8BFpYq",
			Host:     "localhost",
			Port:     "6379",
		}

		os.Setenv("REDIS_PASSWORD", expected.Password)
		os.Setenv("REDIS_HOST", expected.Host)
		os.Setenv("REDIS_PORT", expected.Port)
		defer os.Unsetenv("REDIS_PASSWORD")
		defer os.Unsetenv("REDIS_HOST")
		defer os.Unsetenv("REDIS_PORT")

		v := viper.New()

		redis.SetupEnvRedis(v)

		var conf struct {
			Redis redis.Config `mapstructure:"redis"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, expected.Password, conf.Redis.Password)
		assert.Equal(t, expected.Host, conf.Redis.Host)
		assert.Equal(t, expected.Port, conf.Redis.Port)
	})
}
