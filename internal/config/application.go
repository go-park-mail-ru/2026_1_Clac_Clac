package config

const (
	defaultDebug = true
)

// Конфиг для настройки приложения
type ApplicationConfig struct {
	Debug bool `mapstructure:"debug"`
}

func DefaultApplicationConfig() ApplicationConfig {
	return ApplicationConfig{
		Debug: defaultDebug,
	}
}
