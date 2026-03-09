package config

import (
	"github.com/spf13/viper"
)

// Зануляем все значения по умолчанию
const (
	defaultValue = ""
)

type MailSender struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func DefaultMailSenderConfig() MailSender {
	return MailSender{
		Host:     defaultValue,
		Port:     defaultValue,
		Username: defaultValue,
		Password: defaultValue,
	}
}

func SetDefaultEnvMailSender(v *viper.Viper) {
	v.SetDefault("mail_sender.host", defaultValue)
	v.SetDefault("mail_sender.port", defaultValue)
	v.SetDefault("mail_sender.username", defaultValue)
	v.SetDefault("mail_sender.password", defaultValue)
}
