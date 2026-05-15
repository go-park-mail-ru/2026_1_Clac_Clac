package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel           = DebugLevel
	defaultMaxUploadImageSize = 10 * 1024 * 1024
)

type Application struct {
	LogLevel           string `mapstructure:"log_level"`
	MaxUploadImageSize int64  `mapstructure:"max_upload_image_size"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:           defaultLogLevel,
		MaxUploadImageSize: defaultMaxUploadImageSize,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
