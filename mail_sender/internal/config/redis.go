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
	_                          // used in DefaultRedisConnection
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

		NumberDB:        defaultNumberDB,
		MaxConnections:  defaultMaxRedisConnections,
		MinConnections:  defaultMinRedisConnections,
		PingSleepTime:    defaultPingSleepTimeRedis,
		MaxRetries:      defaultMaxRetriesRedis,
	}
}

func SetupEnvRedisConnection(v *viper.Viper) {
	v.SetDefault("redis.password", defaultValue)
	v.SetDefault("redis.host", defaultValue)
	v.SetDefault("redis.port", defaultValue)
	v.SetDefault("redis.number_db", defaultNumberDB)
	v.SetDefault("redis.min_connections", defaultMinRedisConnections)
	v.SetDefault("redis.max_connections", defaultMaxRedisConnections)
	v.SetDefault("redis.ping_sleep_time", defaultPingSleepTimeRedis)
	v.SetDefault("redis.max_retries", defaultMaxRetriesRedis)

	v.RegisterAlias("redis.password", "redis_password")
	v.RegisterAlias("redis.host", "redis_host")
	v.RegisterAlias("redis.port", "redis_port")
	v.RegisterAlias("redis.number_db", "redis_number_db")
	v.RegisterAlias("redis.min_connections", "redis_min_connections")
	v.RegisterAlias("redis.max_connections", "redis_max_connections")
	v.RegisterAlias("redis.ping_sleep_time", "redis_ping_sleep_time")
	v.RegisterAlias("redis.max_retries", "redis_max_retries")
}
