package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel    = DebugLevel
	defaultMaxFileSize = 10 * 1024 * 1024 // 10 МБайт
)

type Application struct {
	LogLevel    string `mapstructure:"log_level"`
	MaxFileSize int64  `mapstructure:"max_file_size"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:    defaultLogLevel,
		MaxFileSize: defaultMaxFileSize,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
