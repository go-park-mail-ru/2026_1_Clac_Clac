package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App        Application `mapstructure:"app"`
	Engine     Engine      `mapstructure:"engine"`
	MailSender MailSender  `mapstructure:"mail_sender"`
}

func DefaultConfig() Config {
	return Config{
		App:        DefaultApplicationConfig(),
		Engine:     DefaultEngineConfig(),
		MailSender: DefaultMailSenderConfig(),
	}
}

func SetupViper(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetDefaultEnvMailSender(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
