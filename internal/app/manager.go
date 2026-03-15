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

func NewManager(s *Store, mailSenderConf *config.MailSender) *Manager {
	mailSender := mail.NewMailSender(mailSenderConf)

	return &Manager{
		Auth:       auth.NewService(s.Auth, &mailSender, auth.HashPassword, auth.CheckPassword, auth.GenerateSessionID, auth.GeneratorCode),
		Board:      board.NewBoardService(s.Boards),
		Profile:    profile.NewProfileService(s.Profiles),
		MailSender: &mailSender,
	}
}
