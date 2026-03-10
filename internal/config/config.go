package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App        Application `mapstructure:"app"`
	Engine     Engine      `mapstructure:"engine"`
	MailSender MailSender  `mapstructure:"mail_sender"`
	VkOAuth    VkOAuth     `mapstructure:"vk_oauth"`
}

func DefaultConfig() Config {
	return Config{
		App:        DefaultApplicationConfig(),
		Engine:     DefaultEngineConfig(),
		MailSender: DefaultMailSenderConfig(),
		VkOAuth:    DefaultVkOAuthConfig(),
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

	v.SetConfigFile(filepath.Join(configPath, ".env"))
	v.SetConfigType("env")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetupEnvMailSender(v)
	SetupEnvVkOAuth(v)

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
