package config

import (
	"time"
)

const (
	LogInUser    = "login"
	RegisterUser = "register"

	defaultLimit           = 10
	defaultLifeTimeRequest = 50 * time.Millisecond

	safeLimit = 100
	safeTTL   = 3 * time.Second

	defaultCoolDownExpiration = 60
)

type ActionRateLimiters struct {
	Limit  int64         `mapstructure:"limit"`
	Action string        `mapstructure:"action"`
	Window time.Duration `mapstructure:"window"`
	TTL    time.Duration `mapstructure:"ttl"`
}

type RateLimiters struct {
	ClientConfig          `mapstructure:",squash"`
	DBActions             map[string]ActionRateLimiters `mapstructure:"actions"`
	CoolDownExpirationSec int                           `mapstructure:"cool_down_expiration"`
}

func (d *RateLimiters) GetParameters(action string) ActionRateLimiters {
	actionLimit, exist := d.DBActions[action]
	if !exist {
		return ActionRateLimiters{
			Limit: safeLimit,
			TTL:   safeTTL,
		}
	}

	return actionLimit
}

func DefaultActionsRateLimiters() RateLimiters {
	return RateLimiters{
		ClientConfig: DefaultClientConfig(),
		DBActions: map[string]ActionRateLimiters{
			LogInUser: {
				Limit:  defaultLimit,
				Action: LogInUser,
				Window: 1 * time.Minute,
				TTL:    defaultLifeTimeRequest,
			},
			RegisterUser: {
				Limit:  defaultLimit,
				Action: RegisterUser,
				Window: 1 * time.Hour,
				TTL:    defaultLifeTimeRequest,
			},
		},
		CoolDownExpirationSec: defaultCoolDownExpiration,
	}
}
