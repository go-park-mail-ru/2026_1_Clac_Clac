package config

const (
	defaultAddr                    = "localhost:8080"
	defaultWriteTimeout            = 15
	defaultReadTimout              = 15
	defaultIdleTimeout             = 60
	defaultGracefulShutdownTimeout = 15
)

// Конфиг для настройки сервера
type Engine struct {
	Addr                    string `mapstructure:"addr"`
	WriteTimeout            int    `mapstructure:"write_timeout"`
	ReadTimeout             int    `mapstructure:"read_timeout"`
	IdleTimeout             int    `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
}

func DefaultEngineConfig() Engine {
	return Engine{
		Addr:                    defaultAddr,
		WriteTimeout:            defaultWriteTimeout,
		ReadTimeout:             defaultReadTimout,
		IdleTimeout:             defaultIdleTimeout,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}
