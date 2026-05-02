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

	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.number_db", "REDIS_NUMBER_DB")
	v.BindEnv("redis.min_connections", "REDIS_MIN_CONNECTIONS")
	v.BindEnv("redis.max_connections", "REDIS_MAX_CONNECTIONS")
	v.BindEnv("redis.ping_sleep_time", "REDIS_PING_SLEEP_TIME")
	v.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES")
}
