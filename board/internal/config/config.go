package config

import (
	"fmt"
	"strings"

	grpcEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/spf13/viper"
)

type Config struct {
	App      Application         `mapstructure:"app"`
	Engine   grpcEngine.Config   `mapstructure:"engine"`
	Database PostgresConfig      `mapstructure:"database"`
	Redis    RedisConfig         `mapstructure:"redis"`
	S3       S3                  `mapstructure:"s3"`
	Board    Board               `mapstructure:"board"`
	Section  Section             `mapstructure:"section"`
	Card     Card                `mapstructure:"card"`
	Sentry   sentryLogger.Sentry `mapstructure:"sentry"`
	Metrics  Metrics             `mapstructure:"metrics"`
}

func DefaultConfig() Config {
	return Config{
		App:      DefaultApplicationConfig(),
		Board:    DefaultBoardConfig(),
		Section:  DefaultSectionConfig(),
		Card:     DefaultCardConfig(),
		Database: DefaultPostgresConfig(),
		Redis:    DefaultRedisConfig(),
		Sentry:   DefaultSentryConfig(),
		Metrics:  DefaultMetrics(),
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

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetupEnvPostgres(v)
	SetupEnvRedis(v)
	SetupEnvS3(v)
	SetupEnvSentryConfig(v)

	return v, nil
}
