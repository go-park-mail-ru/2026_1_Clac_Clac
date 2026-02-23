package config

const (
	applicationConfigSection = "app"
	defaultDebug             = true
)

// Конфиг для настройки приложения
type ApplicationConfig struct {
	Debug bool `mapstructure:"debug"`
}

func DefaultApplicationConfig() *ApplicationConfig {
	return &ApplicationConfig{
		Debug: defaultDebug,
	}
}

func (c *ApplicationConfig) Section() string {
	return applicationConfigSection
}
