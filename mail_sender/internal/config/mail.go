package config

import (
	"github.com/spf13/viper"
)

const (
	defaultValue = ""
)

type Mail struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func DefaultMailConfig() Mail {
	return Mail{
		Host:     defaultValue,
		Port:     defaultValue,
		Username: defaultValue,
		Password: defaultValue,
	}
}

func SetupEnvMailSender(v *viper.Viper) {
	v.BindEnv("mail.host", "MAIL_SENDER_HOST")
	v.BindEnv("mail.port", "MAIL_SENDER_PORT")
	v.BindEnv("mail.username", "MAIL_SENDER_USERNAME")
	v.BindEnv("mail.password", "MAIL_SENDER_PASSWORD")
}
