package config

const (
	defaultPort = "50051"
)

type GRPCConfig struct {
	Port string `mapstructure:"port"`
}

func DefaultGRPCConfig() GRPCConfig {
	return GRPCConfig{
		Port: defaultPort,
	}
}
