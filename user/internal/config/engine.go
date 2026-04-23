package config

import engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine_grpc"

const (
	defaultAddr                    = ":50053"
	defaultGracefulShutdownTimeout = 15
)

type Engine struct {
	Addr                    string `mapstructure:"addr"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
}

func DefaultEngineConfig() engine.Config {
	return engine.Config{
		Addr:                    defaultAddr,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}
