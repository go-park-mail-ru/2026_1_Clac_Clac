package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel    = DebugLevel
	defaultGRPCMsgSize = 4 * 1024 * 1024 // 4 МБайт
)

type Application struct {
	LogLevel       string `mapstructure:"log_level"`
	MaxMessageSize int64  `mapstructure:"max_message_size"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:       defaultLogLevel,
		MaxMessageSize: defaultGRPCMsgSize,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
