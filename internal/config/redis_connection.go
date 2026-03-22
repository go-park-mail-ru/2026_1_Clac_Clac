package config

import "github.com/spf13/viper"

const (
	defaultNumberDB            = 0
	defaultMaxRedisConnections = 100
	defaultMinRedisConnections = 20
)

type RedisConnection struct {
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`

	NumberDB       int `mapstructure:"number_db"`
	MinConnections int `mapstructure:"min_connections"`
	MaxConnections int `mapstructure:"max_connections"`
}

func DefaultRedisConnection() RedisConnection {
	return RedisConnection{
		Password: defaultValue,
		Host:     defaultValue,
		Port:     defaultValue,

		NumberDB:       defaultNumberDB,
		MaxConnections: defaultMaxRedisConnections,
		MinConnections: defaultMinRedisConnections,
	}
}

func SetupEnvRedisConnection(v *viper.Viper) {
	v.SetDefault("redis.password", defaultValue)
	v.SetDefault("redis.host", defaultValue)
	v.SetDefault("redis.port", defaultValue)

	v.RegisterAlias("redis.password", "redis_password")
	v.RegisterAlias("redis.host", "redis_host")
	v.RegisterAlias("redis.port", "redis_port")
}
