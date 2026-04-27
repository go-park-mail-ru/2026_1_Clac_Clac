package engine

<<<<<<< HEAD
type Config struct {
	Addr                    string `mapstructure:"addr"`
	WriteTimeout            int    `mapstructure:"write_timeout"`
	ReadTimeout             int    `mapstructure:"read_timeout"`
	IdleTimeout             int    `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout int    `mapstructure:"graceful_shutdown_timeout"`
=======
import "time"

type Config struct {
	Addr                    string        `mapstructure:"addr"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
>>>>>>> feat/add-facade
}
