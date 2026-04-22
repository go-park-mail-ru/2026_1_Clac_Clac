package config

const (
	defaultPort = "50052"
)

type GRPCConfig struct {
	Port string `mapstructure:"port"`
}

func DefaultGRPCConfig() GRPCConfig {
	return GRPCConfig{
		Port: defaultPort,
	}
}
