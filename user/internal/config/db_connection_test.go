package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConnectionConfig(t *testing.T) {
	t.Run("test reading database config from env", func(t *testing.T) {
		expected := DatabaseConnection{
			User:     "postgres",
			Password: "Artembobr12345",
			Host:     "localhost",
			Port:     "5434",
			Name:     "nexus_time_manager",
		}

		t.Setenv("DATABASE_USER", expected.User)
		t.Setenv("DATABASE_PASSWORD", expected.Password)
		t.Setenv("DATABASE_HOST", expected.Host)
		t.Setenv("DATABASE_PORT", expected.Port)
		t.Setenv("DATABASE_NAME", expected.Name)

		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()

		SetupEnvDbConnection(v)

		var conf struct {
			Database DatabaseConnection `mapstructure:"database"`
		}

		err := v.Unmarshal(&conf)

		require.NoError(t, err, "viper must not return error")
		assert.Equal(t, expected, conf.Database, "expected other configuration database")
	})
}
