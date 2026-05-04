package config

import (
	"time"
)

const (
	authConfigDefaultMaxLenPassword    = 128
	authConfigDefaultMinLenPassword    = 8
	authConfigSessionLifeTime          = 24 * time.Hour
	authConfigDefaultVKOAuthRedirectTo = "/"
	vkOAuthDefaultValue                = ""
)

type HandlerAuth struct {
	MaxLenPassword    int           `mapstructure:"max_len_password"`
	MinLenPassword    int           `mapstructure:"min_len_password"`
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
			SessionLifetime:   authConfigSessionLifeTime,
			VKOAuthRedirectTo: authConfigDefaultVKOAuthRedirectTo,
		},
		Client: ClientAuth{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
