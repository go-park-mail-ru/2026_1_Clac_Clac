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
	v.SetDefault("csrf.ttl", defaultTTL)
	v.SetDefault("csrf.expire_time_convertation_base", defaultCSRFTokenExpireTimeConvertationBase)
	v.SetDefault("csrf.expire_time_convertation_type_size", defaultCSRFTokenExpireTimeConvertationTypeSize)
	v.SetDefault("csrf.parts_count", defaultPartsCount)

	v.BindEnv("csrf.secret", "CSRF_SECRET")
	v.BindEnv("csrf.ttl", "CSRF_TTL")
	v.BindEnv("csrf.expire_time_convertation_base", "CSRF_EXPIRE_TIME_CONVERTATION_BASE")
	v.BindEnv("csrf.expire_time_convertation_type_size", "CSRF_EXPIRE_TIME_CONVERTATION_TYPE_SIZE")
	v.BindEnv("csrf.parts_count", "CSRF_PARTS_COUNT")
}
