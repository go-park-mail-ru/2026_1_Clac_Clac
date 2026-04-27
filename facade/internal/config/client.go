package config

import "time"

const (
	defaultTimeOut = 5 * time.Second
	defaultRetries = 5
)

type ClientConfig struct {
	Addr    string        `mapstructure:"addr"`
	TimeOut time.Duration `mapstructure:"timeout"`
	Retries int           `mapstructure:"retries"`
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Addr:    defaultAddr,
		TimeOut: defaultTimeOut,
		Retries: defaultRetries,
	}
}
