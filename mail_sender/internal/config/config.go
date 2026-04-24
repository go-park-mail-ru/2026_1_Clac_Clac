package config

import (
	"fmt"
	"strings"

	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine_grpc"
	"github.com/spf13/viper"
)

type Config struct {
	App    Application   `mapstructure:"app"`
	Engine engine.Config `mapstructure:"engine"`
	Mail   Mail          `mapstructure:"mail"`
	Sender Sender        `mapstructure:"sender"`

	RedisConnection RedisConnection `mapstructure:"redis"`
}

func DefaultConfig() Config {
	return Config{
		App:    DefaultApplicationConfig(),
		Engine: DefaultEngineConfig(),
		Mail:   DefaultMailConfig(),
		Sender: DefaultSenderConfig(),

		RedisConnection: DefaultRedisConnection(),
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

	SetupEnvMailSender(v)
	SetupEnvRedisConnection(v)

	return v, nil
}
