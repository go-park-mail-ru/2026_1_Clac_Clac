package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	"github.com/spf13/viper"
)

const (
	defaultWriteTimeout            = 15 * time.Second
	defaultReadTimeout             = 15 * time.Second
	defaultIdleTimeout             = 60 * time.Second
	defaultGracefulShutdownTimeout = 15 * time.Second
)

func DefaultEngineConfig() engine.Config {
	return engine.Config{
		WriteTimeout:            defaultWriteTimeout,
		ReadTimeout:             defaultReadTimeout,
		IdleTimeout:             defaultIdleTimeout,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}

type Config struct {
	App      Application   `mapstructure:"app"`
	Engine   engine.Config `mapstructure:"engine"`
	Broker   BrokerConfig  `mapstructure:"broker"`
	Services Services      `mapstructure:"services"`
}

func DefaultConfig() Config {
	return Config{
		App:      DefaultApplicationConfig(),
		Engine:   DefaultEngineConfig(),
		Broker:   DefaultBrokerConfig(),
		Services: DefaultServicesConfig(),
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

	SetupEnvBroker(v)

	return v, nil
}
