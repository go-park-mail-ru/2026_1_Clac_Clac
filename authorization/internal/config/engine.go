package config

import engineConfig "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"

const (
	defaultAddr                    = ":50054"
	defaultGracefulShutdownTimeout = 15
)

func DefaultEngineConfig() engineConfig.Config {
	return engineConfig.Config{
		Addr:                    defaultAddr,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}
