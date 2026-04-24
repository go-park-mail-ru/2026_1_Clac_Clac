package config

import (
	"fmt"
	"strings"

	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine_grpc"
	"github.com/spf13/viper"
)

type Config struct {
	App          Application        `mapstructure:"app"`
	Engine       engine.Config      `mapstructure:"engine"`
	VkOAuth      VkOAuth            `mapstructure:"vk_oauth"`
	DBConnection DatabaseConnection `mapstructure:"database"`
	S3           S3                 `mapstructure:"s3"`
	S3Avatars    S3Avatars          `mapstructure:"s3_avatars"`

	User User `mapstructure:"user"`
}

func DefaultConfig() Config {
	return Config{
		App:          DefaultApplicationConfig(),
		Engine:       DefaultEngineConfig(),
		VkOAuth:      DefaultVkOAuthConfig(),
		DBConnection: DefaultDBConnectionConfog(),
		S3Avatars:    DefaultS3AvatarsConfig(),
		S3:           DefaultS3Config(),

		User: DefaultUserConfig(),
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
	SetupEnvS3(v)

	return v, nil
}
