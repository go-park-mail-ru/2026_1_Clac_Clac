package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	defaultCSRFSecret                              = ""
	defaultTTL                                     = 24 * time.Hour
	defaultCSRFTokenExpireTimeConvertationBase     = 10
	defaultCSRFTokenExpireTimeConvertationTypeSize = 64
	defaultPartsCount                              = 2
)

type CSRF struct {
	TTL                            time.Duration `mapstructure:"ttl"`
	Secret                         string        `mapstructure:"secret"`
	ExpireTimeConvertationBase     int           `mapstructure:"expire_time_convertation_base"`
	ExpireTimeConvertationTypeSize int           `mapstructure:"expire_time_convertation_type_size"`
	PartsCount                     int           `mapstructure:"parts_count"`
}

func DefaultCSRFConfig() CSRF {
	return CSRF{
		TTL:                            defaultTTL,
		Secret:                         defaultCSRFSecret,
		ExpireTimeConvertationBase:     defaultCSRFTokenExpireTimeConvertationBase,
		ExpireTimeConvertationTypeSize: defaultCSRFTokenExpireTimeConvertationTypeSize,
		PartsCount:                     defaultPartsCount,
	}
}

func SetupEnvCSRFConfig(v *viper.Viper) {
	v.SetDefault("csrf.secret", defaultCSRFSecret)
	v.RegisterAlias("csrf.secret", "csrf_secret")
}
