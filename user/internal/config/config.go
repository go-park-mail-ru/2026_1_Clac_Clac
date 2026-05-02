package config

import (
	"fmt"
	"reflect"
	"strings"

	engine "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpcEngine"
	"github.com/spf13/viper"
)

type Config struct {
	App          Application        `mapstructure:"app"`
	Engine       engine.Config      `mapstructure:"engine"`
	VkOAuth      VkOAuth            `mapstructure:"vk_oauth"`
	DBConnection DatabaseConnection `mapstructure:"database"`
	S3           S3                 `mapstructure:"s3"`
	S3Avatars    S3Avatars          `mapstructure:"s3_avatars"`
}

func DefaultConfig() Config {
	return Config{
		App:          DefaultApplicationConfig(),
		Engine:       DefaultEngineConfig(),
		VkOAuth:      DefaultVkOAuthConfig(),
		DBConnection: DefaultDBConnectionConfog(),
		S3Avatars:    DefaultS3AvatarsConfig(),
		S3:           DefaultS3Config(),
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

	// SetupEnvDbConnection(v)
	// SetupEnvS3(v)
	BindStructKeys(v, Config{})

	return v, nil
}

func BindStructKeys(v *viper.Viper, conf any, parts ...string) {
	bindTypeKeys(v, reflect.TypeOf(conf), parts)
}

func bindTypeKeys(v *viper.Viper, t reflect.Type, parts []string) {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t == nil || t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("mapstructure")

		tag = strings.Split(tag, ",")[0]

		if tag == "-" {
			continue
		}

		if tag == "" {
			tag = strings.ToLower(field.Name)
		}

		currentPath := append(parts[:len(parts):len(parts)], tag)
		fullKey := strings.Join(currentPath, ".")

		fieldType := field.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			bindTypeKeys(v, fieldType, currentPath)
		} else {
			v.BindEnv(fullKey)
		}
	}
}
