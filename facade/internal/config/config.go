package config

import (
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/spf13/viper"
)

type Config struct {
	App      Application         `mapstructure:"app"`
	Engine   engine.Config       `mapstructure:"engine"`
	CORS     CORS                `mapstructure:"cors"`
	CSRF     CSRF                `mapstructure:"csrf"`
	Sentry   sentryLogger.Sentry `mapstructure:"sentry"`
	Services Services            `mapstructure:"services"`
}

func DefaultConfig() Config {
	return Config{
		App:    DefaultApplicationConfig(),
		Engine: DefaultEngineConfig(),
		CORS:   DefaultCORSConfig(),
		CSRF:   DefaultCSRFConfig(),
		Sentry: DefaultSentryConfig(),

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

	engine.SetupEnvEngine(v)
	SetupEnvCORS(v)
	SetupEnvCSRFConfig(v)
	SetupEnvAuth(v)
	SetupEnvSentryConfig(v)

	return v, nil
}
