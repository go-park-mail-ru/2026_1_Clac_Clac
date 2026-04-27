package grpcEngine

type Config struct {
	Addr                    string `mapstructure:"addr"`
	WriteTimeout            int    `mapstructure:"write_timeout"`
	ReadTimeout             int    `mapstructure:"read_timeout"`
	IdleTimeout             int    `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
}
