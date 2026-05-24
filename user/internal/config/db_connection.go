package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultMinPostgresConnections = 2
	defaultMaxPostgresConnections = 10
	deafultMaxConnectionLifetime  = 1 * time.Hour
	defaultMaxHealthCheckPeriod   = 30 * time.Second
	defaultPingSleepTimeDB        = 2 * time.Second
	defaultTimeOut                = 5 * time.Second
	defaultMaxRetriesDB           = 5
	defaultValue                  = ""
)

type DatabaseConnection struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`

	MinConnections        int32         `mapstructure:"min_connections"`
	MaxConnections        int32         `mapstructure:"max_connections"`
	MaxConnectionLifetime time.Duration `mapstructure:"max_connection_lifetime"`
	MaxHealthCheckPeriod  time.Duration `mapstructure:"max_health_check_period"`
	PingSleepTime         time.Duration `mapstructure:"ping_sleep_time"`
	TimeOut               time.Duration `mapstructure:"time_out"`
	MaxRetries            int           `mapstructure:"max_retries"`
}

func DefaultDBConnectionConfog() DatabaseConnection {
	return DatabaseConnection{
		User:     defaultValue,
		Password: defaultValue,
		Host:     defaultValue,
		Port:     defaultValue,
		Name:     defaultValue,

		MinConnections:        defaultMinPostgresConnections,
		MaxConnections:        defaultMaxPostgresConnections,
		MaxConnectionLifetime: deafultMaxConnectionLifetime,
		MaxHealthCheckPeriod:  defaultMaxHealthCheckPeriod,
		PingSleepTime:         defaultPingSleepTimeDB,
		TimeOut:               defaultTimeOut,
		MaxRetries:            defaultMaxRetriesDB,
	}
}

func SetupEnvDbConnection(v *viper.Viper) {
	v.SetDefault("database.user", defaultValue)
	v.SetDefault("database.password", defaultValue)
	v.SetDefault("database.host", defaultValue)
	v.SetDefault("database.port", defaultValue)
	v.SetDefault("database.name", defaultValue)

	v.RegisterAlias("database.user", "database_user")
	v.RegisterAlias("database.password", "database_password")
	v.RegisterAlias("database.host", "database_host")
	v.RegisterAlias("database.port", "database_port")
	v.RegisterAlias("database.name", "database_name")
}
