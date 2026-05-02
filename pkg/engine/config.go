package engine

import (
	"time"
)

type Config struct {
	Addr                    string        `mapstructure:"addr"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}
