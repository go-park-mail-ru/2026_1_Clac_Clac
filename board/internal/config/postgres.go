package config

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/postgres"
	"github.com/spf13/viper"
)

type PostgresConfig struct {
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

func (p PostgresConfig) ToPkg() *postgres.Config {
	return &postgres.Config{
		User:                  p.User,
		Password:              p.Password,
		Host:                  p.Host,
		Port:                  p.Port,
		Name:                  p.Name,
		MinConnections:        p.MinConnections,
		MaxConnections:        p.MaxConnections,
		MaxConnectionLifetime: p.MaxConnectionLifetime,
		MaxHealthCheckPeriod:  p.MaxHealthCheckPeriod,
		PingSleepTime:         p.PingSleepTime,
		TimeOut:               p.TimeOut,
		MaxRetries:            p.MaxRetries,
	}
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		MinConnections:        10,
		MaxConnections:        100,
		MaxConnectionLifetime: time.Hour,
		MaxHealthCheckPeriod:  30 * time.Second,
		PingSleepTime:         2 * time.Second,
		TimeOut:               10 * time.Second,
		MaxRetries:            5,
	}
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

	v.RegisterAlias("database.user", "database_user")
	v.RegisterAlias("database.password", "database_password")
	v.RegisterAlias("database.host", "database_host")
	v.RegisterAlias("database.port", "database_port")
	v.RegisterAlias("database.name", "database_name")
	v.RegisterAlias("database.min_connections", "database_min_connections")
	v.RegisterAlias("database.max_connections", "database_max_connections")
	v.RegisterAlias("database.max_connection_lifetime", "database_max_connection_lifetime")
	v.RegisterAlias("database.max_health_check_period", "database_max_health_check_period")
	v.RegisterAlias("database.ping_sleep_time", "database_ping_sleep_time")
	v.RegisterAlias("database.time_out", "database_time_out")
	v.RegisterAlias("database.max_retries", "database_max_retries")
}
