package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App             Application          `mapstructure:"app"`
	Engine          Engine               `mapstructure:"engine"`
	MailSender      MailSender           `mapstructure:"mail_sender"`
	VkOAuth         VkOAuth              `mapstructure:"vk_oauth"`
	DBConnection    DatabaseConnection   `mapstructure:"database"`
	RedisConnection RedisConnection      `mapstructure:"redis"`
	S3              S3                   `mapstructure:"s3"`
	CORS            CORS                 `mapstructure:"cors"`
	DBRateLimiters  DataBaseRateLimiters `mapstructure:"database_rate_limiters"`
	S3Avatars       S3Avatars            `mapstructure:"s3_avatars"`
	Auth            Auth                 `mapstructure:"auth"`
	Board           Board                `mapstructure:"board"`
	Section         Section              `mapstructure:"section"`
	Profile         Profile              `mapstructure:"profile"`
	Card            Card                 `mapstructture:"card"`
	Appeal          Appeal               `mapstructure:"appeal"`
}

func DefaultConfig() Config {
	return Config{
		App:             DefaultApplicationConfig(),
		Engine:          DefaultEngineConfig(),
		MailSender:      DefaultMailSenderConfig(),
		VkOAuth:         DefaultVkOAuthConfig(),
		DBConnection:    DefaultDBConnectionConfog(),
		RedisConnection: DefaultRedisConnection(),
		S3Avatars:       DefaultS3AvatarsConfig(),
		S3:              DefaultS3Config(),
		Board:           DefaultBoardConfig(),
		CORS:            DefaultCORSConfig(),
		DBRateLimiters:  DefaultActionsRateLimiters(),
		Auth:            DefaultAuthConfig(),
		Profile:         DefaultProfileConfig(),
		Section:         DefaultSectionConfig(),
		Card:            DefaultCardConfig(),
		Appeal:          DefaultAppealConfig(),
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

	SetupEnvMailSender(v)
	SetupEnvVkOAuth(v)
	SetupEnvDbConnection(v)
	SetupEnvRedisConnection(v)
	SetupEnvS3(v)

	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

	return v, nil
}
