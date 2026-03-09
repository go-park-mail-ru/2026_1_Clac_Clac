package config

import "os"

type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
}

func DefaultSMTPSender() SMTPSender {
	return SMTPSender{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}
}
