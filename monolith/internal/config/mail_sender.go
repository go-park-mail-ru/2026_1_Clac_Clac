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

func SetupEnvMailSender(v *viper.Viper) {
	// Надо для того, чтобы viper мог считать переменные из окружения
	v.SetDefault("mail_sender.host", defaultValue)
	v.SetDefault("mail_sender.port", defaultValue)
	v.SetDefault("mail_sender.username", defaultValue)
	v.SetDefault("mail_sender.password", defaultValue)
	// Надо, чтобы viper мог считать .env файл
	v.RegisterAlias("mail_sender.host", "mail_sender_host")
	v.RegisterAlias("mail_sender.port", "mail_sender_port")
	v.RegisterAlias("mail_sender.username", "mail_sender_username")
	v.RegisterAlias("mail_sender.password", "mail_sender_password")
}
