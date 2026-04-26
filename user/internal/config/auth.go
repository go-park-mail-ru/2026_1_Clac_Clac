package config

import (
	"time"
)

const (
	authConfigDefaultLifeTime = 24 * time.Hour
)

type AuthService struct {
	SessionLifetime time.Duration `mapstructure:"session_life_time"`
}

type Auth struct {
	Service AuthService `mapstructure:"service"`
}

func DefaultAuthConfig() Auth {
	return Auth{
		Service: AuthService{
			SessionLifetime: authConfigDefaultLifeTime,
		},
	}
}
