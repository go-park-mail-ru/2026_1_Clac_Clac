package config

import (
	"fmt"
	"strings"

	enginegrpc "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/redis"
	"github.com/spf13/viper"
)

type Config struct {
	App             Application       `mapstructure:"app"`
	Engine          enginegrpc.Config `mapstructure:"engine"`
	Redis           redis.Config      `mapstructure:"redis"`
	RedisConnection RedisConnection   `mapstructure:"-"`
}

func DefaultConfig() Config {
	return Config{
		App:    DefaultApplicationConfig(),
		Engine: DefaultEngineConfig(),
		Redis:  redis.Config{},
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

	redis.SetupEnvRedis(v)

	return v, nil
}
