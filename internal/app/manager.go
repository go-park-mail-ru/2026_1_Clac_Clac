package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/auth/service"
	board "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/board/service"
	card "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/card/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	mail "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/mail_sender/service"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/profile/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/s3"
	section "github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/section/service"
)

type Manager struct {
	Auth       *auth.Service
	Board      *board.Service
	Profile    *profile.Service
	Section    *section.Service
	Card       *card.Service
	MailSender *mail.MailSender
}

func NewManager(s *Store, conf config.Config) *Manager {
	mailSender := mail.NewMailSender(&conf.MailSender)

	authDeps := auth.Tools{
		Rep:    s.Auth,
		Sender: &mailSender,

		Hasher:             auth.HashPassword,
		Checker:            auth.CheckPassword,
		GeneratorID:        auth.GenerateSessionID,
		GeneratorResetCode: auth.GeneratorCode,
		CreaterResetKey:    auth.CreaterResetKey,
		CreaterSessionKey:  auth.CreaterSessionKey,

		CsrfSecret:      conf.Auth.Service.CSRFSecret,
		SessionLifetime: conf.Auth.Service.SessionLifetime,
		CountRetries:    conf.Auth.Service.CountRetries,
	}

	profileConfig := profile.Config{
		GenerateAvatarKey: profile.GenerateAvatarKey,
		BaseURLAvatar:     s3.GetURL(conf.S3.Endpoint, conf.S3.AvatarsBucket),
	}

	return &Manager{
		Auth:       auth.NewService(authDeps),
		Board:      board.NewService(s.Board),
		Profile:    profile.NewService(s.Profile, profileConfig),
		Section:    section.NewService(s.Section),
		Card:       card.NewService(s.Card),
		MailSender: &mailSender,
	}
}
