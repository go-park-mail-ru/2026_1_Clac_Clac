package postgres

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	User                  string        `mapstructure:"user"`
	Password              string        `mapstructure:"password"`
	Host                  string        `mapstructure:"host"`
	Port                  string        `mapstructure:"port"`
	Name                  string        `mapstructure:"name"`
	MinConnections        int32         `mapstructure:"min_connections"`
	MaxConnections        int32         `mapstructure:"max_connections"`
	MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`
	MaxHealthCheckPeriod  time.Duration `mapstructure:"max_health_check_period"`
	PingSleepTime         time.Duration `mapstructure:"ping_sleep_time"`
	TimeOut               time.Duration `mapstructure:"time_out"`
	MaxRetries            int           `mapstructure:"max_retries"`
}

func SetupEnvPostgres(v *viper.Viper) {
	v.SetDefault("database.user", "")
	v.SetDefault("database.password", "")
	v.SetDefault("database.host", "")
	v.SetDefault("database.port", "")
	v.SetDefault("database.name", "")
	v.SetDefault("database.min_connections", int32(2))
	v.SetDefault("database.max_connections", int32(10))
	v.SetDefault("database.max_connection_lifetime", time.Hour)
	v.SetDefault("database.max_health_check_period", 30*time.Second)
	v.SetDefault("database.ping_sleep_time", 2*time.Second)
	v.SetDefault("database.time_out", 5*time.Second)
	v.SetDefault("database.max_retries", 5)

	v.BindEnv("database.user", "DATABASE_USER")
	v.BindEnv("database.password", "DATABASE_PASSWORD")
	v.BindEnv("database.host", "DATABASE_HOST")
	v.BindEnv("database.port", "DATABASE_PORT")
	v.BindEnv("database.name", "DATABASE_NAME")
	v.BindEnv("database.min_connections", "DATABASE_MIN_CONNECTIONS")
	v.BindEnv("database.max_connections", "DATABASE_MAX_CONNECTIONS")
	v.BindEnv("database.max_connection_lifetime", "DATABASE_MAX_CONNECTION_LIFETIME")
	v.BindEnv("database.max_health_check_period", "DATABASE_MAX_HEALTH_CHECK_PERIOD")
	v.BindEnv("database.ping_sleep_time", "DATABASE_PING_SLEEP_TIME")
	v.BindEnv("database.time_out", "DATABASE_TIME_OUT")
	v.BindEnv("database.max_retries", "DATABASE_MAX_RETRIES")
}
