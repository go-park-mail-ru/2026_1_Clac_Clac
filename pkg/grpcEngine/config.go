package enginegrpc

import "github.com/spf13/viper"

type Config struct {
	Addr                    string `mapstructure:"addr"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
}

func SetupEnvGrpcEngine(v *viper.Viper) {
	v.SetDefault("engine.addr", ":50053")
	v.SetDefault("engine.graceful_shutdown_timeout", 15)

	v.RegisterAlias("engine.addr", "GRPCENGINE_ADDR")
	v.RegisterAlias("engine.graceful_shutdown_timeout", "GRPCENGINE_GRACEFUL_SHUTDOWN_TIMEOUT")
}
