package config

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel           = DebugLevel
	defaultMaxTextRequestSize = 10 * 1024        // 10 кБ
	defaultMaxUploadImageSize = 10 * 1024 * 1024 // 10 МБайт
)

type Application struct {
	LogLevel           string `mapstructure:"log_level"`
	MaxTextRequestSize int64  `mapstructure:"max_text_request_size"`
	MaxUploadImageSize int64  `mapstructure:"max_upload_image_size"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:           defaultLogLevel,
		MaxTextRequestSize: defaultMaxTextRequestSize,
		MaxUploadImageSize: defaultMaxUploadImageSize,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
