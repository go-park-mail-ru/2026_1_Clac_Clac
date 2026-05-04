package redis

import (
	"time"

	"github.com/spf13/viper"
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

func SetupEnvRedis(v *viper.Viper) {
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.host", "")
	v.SetDefault("redis.port", "")
	v.SetDefault("redis.number_db", 0)
	v.SetDefault("redis.min_connections", 20)
	v.SetDefault("redis.max_connections", 100)
	v.SetDefault("redis.ping_sleep_time", 2*time.Second)
	v.SetDefault("redis.max_retries", 5)

	v.RegisterAlias("redis.password", "redis_password")
	v.RegisterAlias("redis.host", "redis_host")
	v.RegisterAlias("redis.port", "redis_port")
	v.RegisterAlias("redis.number_db", "redis_number_db")
	v.RegisterAlias("redis.min_connections", "redis_min_connections")
	v.RegisterAlias("redis.max_connections", "redis_max_connections")
	v.RegisterAlias("redis.ping_sleep_time", "redis_ping_sleep_time")
	v.RegisterAlias("redis.max_retries", "redis_max_retries")
}
