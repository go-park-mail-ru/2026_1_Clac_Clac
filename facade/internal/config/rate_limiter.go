package config

import (
	"time"
)

const (
	LogInUser    = "login"
	RegisterUser = "register"

	defaultLimit = 5
)

type ActionRateLimiters struct {
	Limit  int64         `mapstructure:"limit"`
	Action string        `mapstructure:"action"`
	Window time.Duration `mapstructure:"window"`
}

type DataBaseRateLimiters struct {
	DBActions map[string]ActionRateLimiters `mapstructure:"database_actions"`
}

func (d *DataBaseRateLimiters) GetParameters(action string) ActionRateLimiters {
	return d.DBActions[action]
}

func DefaultActionsRateLimiters() DataBaseRateLimiters {
	return DataBaseRateLimiters{
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
	}
}
