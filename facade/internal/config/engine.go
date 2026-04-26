package config

import (
	"time"

	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
)

const (
	defaultAddr                    = "localhost:8081"
	defaultWriteTimeout            = 15 * time.Second
	defaultReadTimout              = 15 * time.Second
	defaultIdleTimeout             = 60 * time.Second
	defaultGracefulShutdownTimeout = 15 * time.Second
)

func DefaultEngineConfig() engine.Config {
	return engine.Config{
		Addr:                    defaultAddr,
		WriteTimeout:            defaultWriteTimeout,
		ReadTimeout:             defaultReadTimout,
		IdleTimeout:             defaultIdleTimeout,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}
