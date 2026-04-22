package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/grpc_engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/postgres"
	"github.com/spf13/viper"
)

type Config struct {
	App      Application        `mapstructure:"app"`
	Engine   grpc_engine.Config `mapstructure:"engine"`
	Database postgres.Config    `mapstructure:"database"`
	S3       S3                 `mapstructure:"s3"`
	Board    Board              `mapstructure:"board"`
	Section  Section            `mapstructure:"section"`
	Card     Card               `mapstructure:"card"`
}

func DefaultConfig() Config {
	return Config{
		App:     DefaultApplicationConfig(),
		Board:   DefaultBoardConfig(),
		Section: DefaultSectionConfig(),
		Card:    DefaultCardConfig(),
	}
}

func SetupViper(configPath string) (*viper.Viper, error) {
	v := viper.New()

	v.AddConfigPath(configPath)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %v", err)
	}

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
