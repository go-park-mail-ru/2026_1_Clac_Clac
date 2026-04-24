package config

import (
	"github.com/spf13/viper"
)

const (
	authConfigDefaultValue          = ""
	authConfigDefaultMaxLenPassword = 128
	authConfigDefaultMinLenPassword = 8

	profileConfigDefaultSiganatureBytes       = 512
	profileConfigDefaultMaxReadBytes          = 5 << 20
	profileConfigDefaultMaxLenNameUser        = 128
	profileConfigDefaultMaxLenDescriptionUser = 500
)

type UserHandler struct {
	MaxLenPassword int `mapstructure:"max_len_password"`
	MinLenPassword int `mapstructure:"min_len_password"`

	SiganatureTypeBytes   int   `mapstructure:"signature_type_bytes"`
	MaxReadBytes          int64 `mapstructure:"max_read_bytes"`
	MaxLenNameUser        int   `mapstructure:"max_len_name_user"`
	MaxLenDescriptionUser int   `mapstructure:"max_len_description_user"`
}

type User struct {
	Handler UserHandler `mapstructure:"handler"`
}

func DefaultUserConfig() User {
	return User{
		Handler: UserHandler{
			MaxLenPassword:        authConfigDefaultMaxLenPassword,
			MinLenPassword:        authConfigDefaultMinLenPassword,
			SiganatureTypeBytes:   profileConfigDefaultSiganatureBytes,
			MaxReadBytes:          profileConfigDefaultMaxReadBytes,
			MaxLenNameUser:        profileConfigDefaultMaxLenNameUser,
			MaxLenDescriptionUser: profileConfigDefaultMaxLenDescriptionUser,
		},
	}
}

func SetupEnvDUserConfig(v *viper.Viper) {
	v.SetDefault("auth.service.csrf_secret", authConfigDefaultValue)
	v.RegisterAlias("auth.service.csrf_secret", "auth_service_csrf_secret")
}
