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
	v.SetDefault("mail.host", "")
	v.SetDefault("mail.port", "")
	v.SetDefault("mail.username", "")
	v.SetDefault("mail.password", "")

	v.RegisterAlias("mail.host", "mail_sender_host")
	v.RegisterAlias("mail.port", "mail_sender_port")
	v.RegisterAlias("mail.username", "mail_sender_username")
	v.RegisterAlias("mail.password", "mail_sender_password")
}
