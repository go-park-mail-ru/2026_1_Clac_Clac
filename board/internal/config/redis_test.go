package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/board/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConfig(t *testing.T) {
	t.Run("test reading redis config from env", func(t *testing.T) {
		expected := config.RedisConfig{
			Password:       "password123",
			Host:          "localhost",
			Port:          "6379",
			NumberDB:      1,
			MinConnections: 10,
			MaxConnections: 50,
		}

		t.Setenv("REDIS_PASSWORD", expected.Password)
		t.Setenv("REDIS_HOST", expected.Host)
		t.Setenv("REDIS_PORT", expected.Port)
		t.Setenv("REDIS_NUMBER_DB", "1")
		t.Setenv("REDIS_MIN_CONNECTIONS", "10")
		t.Setenv("REDIS_MAX_CONNECTIONS", "50")

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvRedis(v)

		var conf struct {
			Redis config.RedisConfig `mapstructure:"redis"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, expected, conf.Redis)
	})
}