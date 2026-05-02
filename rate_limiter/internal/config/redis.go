package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultNumberDB            = 0
	defaultMaxRedisConnections = 100
	defaultMinRedisConnections = 20
	defaultPingSleepTimeRedis  = 2 * time.Second
	defaultMaxRetriesRedis     = 5
)

const (
	defaultValue = ""
)

type RedisConnection struct {
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`

	NumberDB       int           `mapstructure:"number_db"`
	MinConnections int           `mapstructure:"min_connections"`
	MaxConnections int           `mapstructure:"max_connections"`
	PingSleepTime  time.Duration `mapstructure:"ping_sleep_time"`
	MaxRetries     int           `mapstructure:"max_retries"`
}

func DefaultRedisConnection() RedisConnection {
	return RedisConnection{
		Password: defaultValue,
		Host:     defaultValue,
		Port:     defaultValue,

		NumberDB:       defaultNumberDB,
		MaxConnections: defaultMaxRedisConnections,
		MinConnections: defaultMinRedisConnections,
		PingSleepTime:  defaultPingSleepTimeRedis,
		MaxRetries:     defaultMaxRetriesRedis,
	}
}

func SetupEnvRedisConnection(v *viper.Viper) {
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
}
