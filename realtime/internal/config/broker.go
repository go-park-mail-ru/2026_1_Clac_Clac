package config

import (
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/spf13/viper"
)

type BrokerConfig struct {
	Password       string        `mapstructure:"password"`
	Host           string        `mapstructure:"host"`
	Port           string        `mapstructure:"port"`
	NumberDB       int           `mapstructure:"number_db"`
	MinConnections int           `mapstructure:"min_connections"`
	MaxConnections int           `mapstructure:"max_connections"`
	PingSleepTime  time.Duration `mapstructure:"ping_sleep_time"`
	MaxRetries     int           `mapstructure:"max_retries"`
}

func (b BrokerConfig) ToPkg() redis.Config {
	return redis.Config{
		Password:       b.Password,
		Host:           b.Host,
		Port:           b.Port,
		NumberDB:       b.NumberDB,
		MinConnections: b.MinConnections,
		MaxConnections: b.MaxConnections,
		PingSleepTime:  b.PingSleepTime,
		MaxRetries:     b.MaxRetries,
	}
}

func DefaultBrokerConfig() BrokerConfig {
	return BrokerConfig{
		NumberDB:        0,
		MinConnections: 20,
		MaxConnections: 100,
		PingSleepTime:  2 * time.Second,
		MaxRetries:     5,
	}
}

func SetupEnvBroker(v *viper.Viper) {
	v.SetDefault("broker.password", os.Getenv("REDIS_PASSWORD"))
	v.SetDefault("broker.host", "")
	v.SetDefault("broker.port", "")
	v.SetDefault("broker.number_db", 0)
	v.SetDefault("broker.min_connections", 20)
	v.SetDefault("broker.max_connections", 100)
	v.SetDefault("broker.ping_sleep_time", 2*time.Second)
	v.SetDefault("broker.max_retries", 5)

	v.RegisterAlias("broker.password", "broker_password")
	v.RegisterAlias("broker.host", "broker_host")
	v.RegisterAlias("broker.port", "broker_port")
	v.RegisterAlias("broker.number_db", "broker_number_db")
	v.RegisterAlias("broker.min_connections", "broker_min_connections")
	v.RegisterAlias("broker.max_connections", "broker_max_connections")
	v.RegisterAlias("broker.ping_sleep_time", "broker_ping_sleep_time")
	v.RegisterAlias("broker.max_retries", "broker_max_retries")
}
