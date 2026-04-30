package config

type MailSender struct {
	ClientConfig `mapstructure:",squash"`
}
