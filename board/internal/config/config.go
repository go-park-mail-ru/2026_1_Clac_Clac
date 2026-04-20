package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App            Application          `mapstructure:"app"`
	Engine         Engine               `mapstructure:"engine"`
	VkOAuth        VkOAuth              `mapstructure:"vk_oauth"`
	S3             S3                   `mapstructure:"s3"`
	CORS           CORS                 `mapstructure:"cors"`
	DBRateLimiters DataBaseRateLimiters `mapstructure:"database_rate_limiters"`
	S3Avatars      S3Avatars            `mapstructure:"s3_avatars"`
	Auth           Auth                 `mapstructure:"auth"`
	Board          Board                `mapstructure:"board"`
	Section        Section              `mapstructure:"section"`
	Profile        Profile              `mapstructure:"profile"`
	Card           Card                 `mapstructture:"card"`
}

func DefaultConfig() Config {
	return Config{
		App:            DefaultApplicationConfig(),
		Engine:         DefaultEngineConfig(),
		VkOAuth:        DefaultVkOAuthConfig(),
		S3Avatars:      DefaultS3AvatarsConfig(),
		S3:             DefaultS3Config(),
		Board:          DefaultBoardConfig(),
		CORS:           DefaultCORSConfig(),
		DBRateLimiters: DefaultActionsRateLimiters(),
		Auth:           DefaultAuthConfig(),
		Profile:        DefaultProfileConfig(),
		Section:        DefaultSectionConfig(),
		Card:           DefaultCardConfig(),
	}
}

func SetupViper(configPath string) (*viper.Viper, error) {
	v := viper.New()

	v.AddConfigPath(configPath)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	// v.SetConfigFile(filepath.Join(configPath, ".env"))
	// v.SetConfigType("env")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetupEnvVkOAuth(v)
	SetupEnvS3(v)

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
