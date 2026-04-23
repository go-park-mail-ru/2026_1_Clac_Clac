package enginegrpc

type Config struct {
	Addr                    string `mapstructure:"addr"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
}
