package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel = DebugLevel
)

// Конфиг для настройки приложения
type Application struct {
	LogLevel string `mapstructure:"log_level"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel: defaultLogLevel,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
