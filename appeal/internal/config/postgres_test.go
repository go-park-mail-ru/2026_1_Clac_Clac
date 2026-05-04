package config_test

import (
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/appeal/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresConfig(t *testing.T) {
	t.Run("test reading postgres config from env", func(t *testing.T) {
		expected := config.PostgresConfig{
			User:                  "postgres",
			Password:              "password123",
			Host:                 "localhost",
			Port:                 "5432",
			Name:                 "nexus_db",
			MinConnections:        5,
			MaxConnections:       20,
			MaxRetries:           3,
		}

		t.Setenv("DATABASE_USER", expected.User)
		t.Setenv("DATABASE_PASSWORD", expected.Password)
		t.Setenv("DATABASE_HOST", expected.Host)
		t.Setenv("DATABASE_PORT", expected.Port)
		t.Setenv("DATABASE_NAME", expected.Name)
		t.Setenv("DATABASE_MIN_CONNECTIONS", "5")
		t.Setenv("DATABASE_MAX_CONNECTIONS", "20")
		t.Setenv("DATABASE_MAX_RETRIES", "3")

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		config.SetupEnvPostgres(v)

		var conf struct {
			Database config.PostgresConfig `mapstructure:"database"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, expected, conf.Database)
	})
}