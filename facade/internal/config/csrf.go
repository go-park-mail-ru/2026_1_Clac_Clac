package config

import "github.com/spf13/viper"

const (
	defaultCSRFSecret = ""
)

type CSRF struct {
	Secret string `mapstructure:"secret"`
}

func DefaultCSRFConfig() CSRF {
	return CSRF{
		Secret: defaultCSRFSecret,
	}
}

func SetupEnvCSRFConfig(v *viper.Viper) {
	v.SetDefault("csrf.secret", defaultCSRFSecret)
	v.RegisterAlias("csrf.secret", "csrf_secret")
}
