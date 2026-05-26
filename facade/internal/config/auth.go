package config

import (
	"time"
)

const (
	authConfigDefaultMaxLenPassword    = 128
	authConfigDefaultMinLenPassword    = 8
	authConfigDefaultMaxLenNameUser    = 128
	authConfigSessionLifeTime          = 24 * time.Hour
	authConfigDefaultVKOAuthRedirectTo = "/"
)

type HandlerAuth struct {
	MaxLenPassword    int           `mapstructure:"max_len_password"`
	MinLenPassword    int           `mapstructure:"min_len_password"`
	MaxLenNameUser    int           `mapstructure:"max_len_name_user"`
	SessionLifetime   time.Duration `mapstructure:"session_life_time"`
	VKOAuthRedirectTo string        `mapstructure:"vk_oauth_redirect_to"`
}

type ClientAuth struct {
	ClientConfig `mapstructure:",squash"`
}

type Auth struct {
	Handler HandlerAuth `mapstructure:"handler"`
	Client  ClientAuth  `mapstructure:"client"`
}

func DefaultAuthConfig() Auth {
	return Auth{
		Handler: HandlerAuth{
			MaxLenPassword:    authConfigDefaultMaxLenPassword,
			MinLenPassword:    authConfigDefaultMinLenPassword,
			MaxLenNameUser:    authConfigDefaultMaxLenNameUser,
			SessionLifetime:   authConfigSessionLifeTime,
			VKOAuthRedirectTo: authConfigDefaultVKOAuthRedirectTo,
		},
		Client: ClientAuth{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
