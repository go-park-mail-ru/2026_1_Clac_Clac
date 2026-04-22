package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/s3"
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/config"
	profile "github.com/go-park-mail-ru/2026_1_Clac_Clac/user/internal/profile/service"
)

type Manager struct {
	Auth    *auth.Service
	Profile *profile.Service
}

func NewManager(s *Store, conf config.Config) *Manager {
	authDeps := auth.Tools{
		Rep: s.Auth,

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
		Auth:    auth.NewService(authDeps),
		Profile: profile.NewService(s.Profile, profileConfig),
	}
}
