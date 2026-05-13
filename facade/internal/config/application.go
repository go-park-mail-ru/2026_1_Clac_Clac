package config

import "time"

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
)

const (
	defaultLogLevel           = DebugLevel
	defaultMaxTextRequestSize = 10 * 1024        // 10 кБ
	defaultMaxUploadImageSize = 10 * 1024 * 1024 // 10 МБайт
	defaultMaxFileSize        = 10 * 1024 * 1024
	defaultRequestTimeout     = 5 * time.Second
)

type Application struct {
	LogLevel           string        `mapstructure:"log_level"`
	MaxTextRequestSize int64         `mapstructure:"max_text_request_size"`
	MaxUploadImageSize int64         `mapstructure:"max_upload_image_size"`
	MaxFileSize        int64         `mapstructure:"max_file_size"`
	RequestTimeout     time.Duration `mapstructure:"request_timeout"`
}

func DefaultApplicationConfig() Application {
	return Application{
		LogLevel:           defaultLogLevel,
		MaxTextRequestSize: defaultMaxTextRequestSize,
		MaxUploadImageSize: defaultMaxUploadImageSize,
		MaxFileSize:        defaultMaxFileSize,
		RequestTimeout:     defaultRequestTimeout,
	}
}

func IsDebug(level string) bool {
	return level == DebugLevel
}
