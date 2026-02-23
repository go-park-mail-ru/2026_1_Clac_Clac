package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config interface {
	Section() string
}

func ReadWithViper(v *viper.Viper, conf Config) error {
	if err := v.Sub(conf.Section()).Unmarshal(conf); err != nil {
		return fmt.Errorf("viper.Unmarshal: %w", err)
	}

	return nil
}
