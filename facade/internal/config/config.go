package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	defaultMonolitURL    = "http://127.0.0.1:8081"
	defaultMailSenderURL = "127.0.0.1:50051"
)

type Config struct {
	App             Application          `mapstructure:"app"`
	Engine          Engine               `mapstructure:"engine"`
	RedisConnection RedisConnection      `mapstructure:"redis"`
	CORS            CORS                 `mapstructure:"cors"`
	DBRateLimiters  DataBaseRateLimiters `mapstructure:"database_rate_limiters"`

	MailSenderURL string `mapstructure:"mail_sender_url"`
	MonolithURL   string `mapstructure:"monolith_url"`
}

func DefaultConfig() Config {
	return Config{
		App:             DefaultApplicationConfig(),
		Engine:          DefaultEngineConfig(),
		RedisConnection: DefaultRedisConnection(),
		CORS:            DefaultCORSConfig(),
		DBRateLimiters:  DefaultActionsRateLimiters(),

		MailSenderURL: defaultMailSenderURL,
		MonolithURL:   defaultMonolitURL,
	}
}

func SetupViper(configPath string) (*viper.Viper, error) {
	v := viper.New()

	v.AddConfigPath(configPath)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	// v.SetConfigFile(filepath.Join(configPath, ".env"))
	// v.SetConfigType("env")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetupEnvRedisConnection(v)

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
