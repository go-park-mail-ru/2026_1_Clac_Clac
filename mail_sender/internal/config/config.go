package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App        Application `mapstructure:"app"`
	GRPC       GRPCConfig  `mapstructure:"grpc"`
	MailSender MailSender  `mapstructure:"mail_sender"`
}

func DefaultConfig() Config {
	return Config{
		App:        DefaultApplicationConfig(),
		GRPC:       DefaultGRPCConfig(),
		MailSender: DefaultMailSenderConfig(),
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

	return v, nil
}
