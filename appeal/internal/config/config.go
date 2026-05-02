package config

import (
	"fmt"
	"strings"

	grpcEngine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/postgres"
	"github.com/spf13/viper"
)

type Config struct {
	App      Application       `mapstructure:"app"`
	Engine   grpcEngine.Config `mapstructure:"engine"`
	Database postgres.Config   `mapstructure:"database"`
	S3       S3                `mapstructure:"s3"`
	Appeal   Appeal            `mapstructure:"appeal"`
}

func DefaultConfig() Config {
	return Config{
		App:      DefaultApplicationConfig(),
		Appeal:   DefaultAppealConfig(),
		Database: postgres.Config{},
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

	postgres.SetupEnvPostgres(v)
	grpcEngine.SetupEnvGrpcEngine(v)
	SetupEnvS3(v)

	return v, nil
}
