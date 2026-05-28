package config

import "time"

type ClientAuth struct {
	ClientConfig `mapstructure:",squash"`
}

type Auth struct {
	SessionLifetime time.Duration `mapstructure:"session_lifetime"`
	Client          ClientAuth    `mapstructure:"client"`
}

func DefaultAuthConfig() Auth {
	return Auth{
		SessionLifetime: 24 * time.Hour,
		Client: ClientAuth{
			ClientConfig: DefaultClientConfig(),
		},
	}
}
