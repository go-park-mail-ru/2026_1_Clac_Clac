package config

import "github.com/spf13/viper"

const (
	defaultCorsCredentials = "false"
	defaultCorsOrigin      = "localhost"
	defaultCorsMethods     = "GET,POST,OPTIONS"
	defaultCorsHeaders     = ""
	defaultCorsMaxAge      = "60"
)

type CORS struct {
	Credentials string `mapstructure:"credentials"`
	Origin      string `mapstructure:"origin"`
	Methods     string `mapstructure:"methods"`
	Headers     string `mapstructure:"headers"`
	MaxAge      string `mapstructure:"max_age"`
}

func DefaultCORSConfig() CORS {
	return CORS{
		Credentials: defaultCorsCredentials,
		Origin:      defaultCorsOrigin,
		Methods:     defaultCorsMethods,
		Headers:     defaultCorsHeaders,
		MaxAge:      defaultCorsMaxAge,
	}
}

func SetupEnvCORS(v *viper.Viper) {
	v.SetDefault("cors.origin", defaultCorsOrigin)
	v.RegisterAlias("cors.origin", "cors_origin")
}
