package db

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
		Password: "",
		Host:     "",
		Port:     "",

		NumberDB:       defaultNumberDB,
		MaxConnections: defaultMaxRedisConnections,
		MinConnections: defaultMinRedisConnections,
		PingSleepTime:  defaultPingSleepTimeRedis,
		MaxRetries:     defaultMaxRetriesRedis,
	}
}

func SetupEnvRedisConnection(v *viper.Viper) {
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.host", "")
	v.SetDefault("redis.port", "")

	v.RegisterAlias("redis.password", "redis_password")
	v.RegisterAlias("redis.host", "redis_host")
	v.RegisterAlias("redis.port", "redis_port")
}
