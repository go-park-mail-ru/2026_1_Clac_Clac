package logger

type Sentry struct {
	DSN              string            `mapstructure:"dsn"`
	Environment      string            `mapstructure:"environment"`
	Release          string            `mapstructure:"release"`
	ServiceName      string            `mapstructure:"service_name"`
	Tags             map[string]string `mapstructure:"tags"`
	TracesSampleRate float64           `mapstructure:"traces_sample_rate"`
}
