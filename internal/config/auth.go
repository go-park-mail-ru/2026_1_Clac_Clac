package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	authConfigDefaultValue          = ""
	authConfigDefaultLifeTime       = 24 * time.Hour
	authConfigDefaultCountRetries   = 3
	authConfigDefaultMaxLenPassword = 128
	authConfigDefaultMinLenPassword = 8
)

type AuthHandler struct {
	MaxLenPassword  int           `mapstructure:"max_len_password"`
	MinLenPassword  int           `mapstructure:"min_len_password"`
	SessionLifetime time.Duration `mapstructure:"session_life_time"`
}

type AuthService struct {
	CSRFSecret      string        `mapstructure:"csrf_secret"`
	SessionLifetime time.Duration `mapstructure:"session_life_time"`
	CountRetries    int           `mapstructure:"count_retries"`
}

type Auth struct {
	Handler AuthHandler `mapstructure:"handler"`
	Service AuthService `mapstructure:"service"`
}

func DefaultAuthConfig() Auth {
	return Auth{
		Handler: AuthHandler{
			MaxLenPassword:  authConfigDefaultMaxLenPassword,
			MinLenPassword:  authConfigDefaultMinLenPassword,
			SessionLifetime: authConfigDefaultLifeTime,
		},
		Service: AuthService{
			CSRFSecret:      authConfigDefaultValue,
			SessionLifetime: authConfigDefaultLifeTime,
			CountRetries:    authConfigDefaultCountRetries,
		},
	}
}

func SetupEnvAuthConfig(v *viper.Viper) {
	v.SetDefault("auth.service.csrf_secret", authConfigDefaultValue)
	v.RegisterAlias("auth.service.csrf_secret", "auth_service_csrf_secret")
}
