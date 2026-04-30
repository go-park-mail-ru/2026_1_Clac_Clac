package config

import (
	"time"
)

const (
	LogInUser    = "login"
	RegisterUser = "register"

	defaultLimit              = 5
	defaultCoolDownExpiration = 60
)

type ActionRateLimiters struct {
	Limit  int64         `mapstructure:"limit"`
	Action string        `mapstructure:"action"`
	Window time.Duration `mapstructure:"window"`
}

type RateLimiters struct {
	ClientConfig          `mapstructure:",squash"`
	DBActions             map[string]ActionRateLimiters `mapstructure:"actions"`
	CoolDownExpirationSec int                           `mapstructure:"cool_down_expiration"`
}

func (d *RateLimiters) GetParameters(action string) ActionRateLimiters {
	return d.DBActions[action]
}

func DefaultActionsRateLimiters() RateLimiters {
	return RateLimiters{
		ClientConfig: DefaultClientConfig(),
		DBActions: map[string]ActionRateLimiters{
			LogInUser: {
				Limit:  defaultLimit,
				Action: LogInUser,
				Window: 1 * time.Minute,
			},
			RegisterUser: {
				Limit:  defaultLimit,
				Action: RegisterUser,
				Window: 1 * time.Hour,
			},
		},
		CoolDownExpirationSec: defaultCoolDownExpiration,
	}
}
