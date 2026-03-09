package mail

import (
	"fmt"
	"net/smtp"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
)

type MailSender struct {
	host     string
	port     string
	username string
	password string
}

func NewMailSender(conf *config.MailSender) MailSender {
	return MailSender{
		host:     conf.Host,
		port:     conf.Port,
		username: conf.Username,
		password: conf.Password,
	}
}

func (ms *MailSender) SendLetter(to string, subject string, htmlBody string) error {
	header := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
		"\r\n", ms.username, to, subject)

	msg := []byte(header + htmlBody)

	auth := smtp.PlainAuth("", ms.username, ms.password, ms.host)

	addr := fmt.Sprintf("%s:%s", ms.host, ms.port)
	err := smtp.SendMail(addr, auth, ms.username, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("smtp.SendMail failed: %w", err)
	}

	return nil
}
