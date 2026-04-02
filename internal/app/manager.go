package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	mail "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/mail_sender/service"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service"
)

type Manager struct {
	Auth       *auth.Service
	Board      *board.Service
	Profile    *profile.Service
	MailSender *mail.MailSender
}

func NewManager(s *Store, conf config.Config) *Manager {
	mailSender := mail.NewMailSender(&conf.MailSender)

	return &Manager{
		Auth: auth.NewFromConfig(auth.AuthServiceConfig{
			AuthRepository:     s.Auth,
			EmailSender:        &mailSender,
			Hasher:             auth.HashPassword,
			Checker:            auth.CheckPassword,
			IdGenerator:        auth.GenerateSessionID,
			ResetCodeGenerator: auth.GeneratorCode,
			CSRFSecret:         conf.Auth.CSRFSecret,
		}),
		Board:      board.NewService(s.Boards),
		Profile:    profile.NewService(s.Profiles),
		MailSender: &mailSender,
	}
}
