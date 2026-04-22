package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App             Application        `mapstructure:"app"`
	GRPC            GRPCConfig         `mapstructure:"grpc"`
	VkOAuth         VkOAuth            `mapstructure:"vk_oauth"`
	DBConnection    DatabaseConnection `mapstructure:"database"`
	RedisConnection RedisConnection    `mapstructure:"redis"`
	S3              S3                 `mapstructure:"s3"`
	S3Avatars       S3Avatars          `mapstructure:"s3_avatars"`

	Auth    Auth    `mapstructure:"auth"`
	Profile Profile `mapstructure:"profile"`
}

func DefaultConfig() Config {
	return Config{
		App:             DefaultApplicationConfig(),
		GRPC:            DefaultGRPCConfig(),
		VkOAuth:         DefaultVkOAuthConfig(),
		DBConnection:    DefaultDBConnectionConfog(),
		RedisConnection: DefaultRedisConnection(),
		S3Avatars:       DefaultS3AvatarsConfig(),
		S3:              DefaultS3Config(),

		Auth:    DefaultAuthConfig(),
		Profile: DefaultProfileConfig(),
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

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	SetupEnvDbConnection(v)
	// SetupEnvRedisConnection(v)
	SetupEnvS3(v)

	return v, nil
}
