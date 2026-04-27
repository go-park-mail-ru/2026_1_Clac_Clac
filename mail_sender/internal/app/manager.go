package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/config"
	mail "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/mail/service"
	sender "github.com/go-park-mail-ru/2026_1_Clac_Clac/mail_sender/internal/sender/service"
)

type Manager struct {
	Sender *sender.Service
}

func NewManager(s *Store, conf config.Config) *Manager {
	mail := mail.NewMailSender(&conf.Mail)

	senderTools := sender.Tools{
		GeneratorResetCode: sender.GeneratorCode,
		CreatorResetKey:    sender.CreatorResetKey,
	}

	senderConfig := sender.Config{
		CountRetries:       conf.Sender.Service.CountRetries,
		LifeTimeResetToken: conf.Sender.Service.LifeTimeResetToken,
		SleepTime:          conf.Sender.Service.SleepTime,
	}

	return &Manager{
		Sender: sender.NewService(s.Sender, mail, senderConfig, senderTools),
	}
}
