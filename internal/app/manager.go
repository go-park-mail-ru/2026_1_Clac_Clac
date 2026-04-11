package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	mail "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/mail_sender/service"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
)

type Manager struct {
	Auth       *auth.Service
	Board      *board.Service
	Profile    *profile.Service
	MailSender *mail.MailSender
}

func NewManager(s *Store, conf config.Config) *Manager {
	mailSender := mail.NewMailSender(&conf.MailSender)
	baseURLAvatar := s3.GenerateBaseURL(conf.S3.AvatarsBucket, conf.S3.Endpoint)

	return &Manager{
		Auth:       auth.NewService(s.Auth, &mailSender, auth.HashPassword, auth.CheckPassword, auth.GenerateSessionID, auth.GeneratorCode, conf.Auth.CSRFSecret, auth.CreaterResetKey, auth.CreaterSessionKey),
		Board:      board.NewService(s.Boards),
		Profile:    profile.NewService(s.Profiles, profile.GenerateAvatarKey, baseURLAvatar),
		MailSender: &mailSender,
	}
}
