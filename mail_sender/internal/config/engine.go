package config

import engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"

const (
	defaultAddr                    = ":50052"
	defaultGracefulShutdownTimeout = 15
)

func DefaultEngineConfig() engine.Config {
	return engine.Config{
		Addr:                    defaultAddr,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}
