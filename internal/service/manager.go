package service

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/repository"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/board"
	mail "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/mail_sender"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/service/profile"
)

type Manager struct {
	Auth       *auth.AuthUserService
	Board      *board.BoardService
	Profile    *profile.ProfileService
	MailSender *mail.MailSender
}

func NewManager(s *repository.Store, mailSenderConf *config.MailSender) *Manager {
	mailSender := mail.NewMailSender(mailSenderConf)

	return &Manager{
		Auth:       auth.NewAuthService(s.Auth, &mailSender, auth.HashPassword, auth.CheckPassword, auth.GenerateSessionID, auth.GeneratorCode),
		Board:      board.NewBoardService(s.Boards),
		Profile:    profile.NewProfileService(s.Profiles),
		MailSender: &mailSender,
	}
}
