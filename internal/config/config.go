package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	AppConfig    ApplicationConfig `mapstructure:"app"`
	EngineConfig EngineConfig      `mapstructure:"http"`
}

func DefaultConfig() Config {
	return Config{
		AppConfig:    DefaultApplicationConfig(),
		EngineConfig: DefaultEngineConfig(),
	}
}

// Настройка viper
func SetupViper(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
