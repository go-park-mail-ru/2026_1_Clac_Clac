package postgres

import (
	"time"
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
