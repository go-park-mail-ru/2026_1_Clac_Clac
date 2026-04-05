package config

import (
	"github.com/spf13/viper"
)

const (
	authConfigDefaultValue = ""
)

type Auth struct {
	CSRFSecret string `mapstructure:"csrf_secret"`
}

func DefaultAuthConfig() Auth {
	return Auth{
		CSRFSecret: authConfigDefaultValue,
	}
}

func SetupEnvAuthConfig(v *viper.Viper) {
	v.SetDefault("auth.csrf_secret", authConfigDefaultValue)
	v.RegisterAlias("auth.csrf_secret", "auth_csrf_secret")
}
