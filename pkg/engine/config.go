package engine

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Addr                    string        `mapstructure:"addr"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

func SetupEnvEngine(v *viper.Viper) {
	v.SetDefault("engine.addr", "localhost:8081")
	v.SetDefault("engine.write_timeout", 15*time.Second)
	v.SetDefault("engine.read_timeout", 15*time.Second)
	v.SetDefault("engine.idle_timeout", 60*time.Second)
	v.SetDefault("engine.graceful_shutdown_timeout", 15*time.Second)

	v.RegisterAlias("engine.addr", "ENGINE_ADDR")
	v.RegisterAlias("engine.write_timeout", "ENGINE_WRITE_TIMEOUT")
	v.RegisterAlias("engine.read_timeout", "ENGINE_READ_TIMEOUT")
	v.RegisterAlias("engine.idle_timeout", "ENGINE_IDLE_TIMEOUT")
	v.RegisterAlias("engine.graceful_shutdown_timeout", "ENGINE_GRACEFUL_SHUTDOWN_TIMEOUT")
}
