package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel           = DebugLevel
	defaultMaxTextRequestSize = 10 * 1024 // 10 кБ
)

type Application struct {
	LogLevel           string `mapstructure:"log_level"`
	MaxTextRequestSize int64  `mapstructure:"max_text_request_size"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:           defaultLogLevel,
		MaxTextRequestSize: defaultMaxTextRequestSize,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
