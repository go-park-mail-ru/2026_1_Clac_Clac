package config

import (
	"time"
)

const (
	defaultCountRetries        = 5
	defaultLifeTimeResetToken  = 15 * time.Minute
	defaultSleepTimeResetToken = 2 * time.Second
)

type SenderService struct {
	CountRetries       int           `mapstructure:"count_retries"`
	LifeTimeResetToken time.Duration `mapstructure:"life_time_reset_token"`
	SleepTime          time.Duration `mapstructure:"sleep_time"`
}

type Sender struct {
	Service SenderService `mapstructure:"service"`
}

func DefaultSenderConfig() Sender {
	return Sender{
		Service: SenderService{
			CountRetries:       defaultCountRetries,
			LifeTimeResetToken: defaultLifeTimeResetToken,
			SleepTime:          defaultSleepTimeResetToken,
		},
	}
}
