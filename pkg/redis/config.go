package redis

import (
	"time"
)

type Config struct {
	Password       string        `mapstructure:"password"`
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	NumberDB       int           `mapstructure:"number_db"`
	MinConnections int           `mapstructure:"min_connections"`
	MaxConnections int           `mapstructure:"max_connections"`
	PingSleepTime  time.Duration `mapstructure:"ping_sleep_time"`
	MaxRetries     int           `mapstructure:"max_retries"`
}
