package config

const (
	defaultMetricsPort = ":9091"
)

type Metrics struct {
	MetricsPort string `mapstructure:"metrics_port"`
}

func DefaultMetrics() Metrics {
	return Metrics{
		MetricsPort: defaultMetricsPort,
	}
}
