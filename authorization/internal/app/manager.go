package app

import (
	auth "github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/auth/service"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/authorization/internal/config"
)

type Manager struct {
	Auth *auth.Service
}

func NewManager(s *Store, conf config.Config) *Manager {
	authTools := auth.Tools{
		GeneratorSessionID: auth.GenerateSessionKey,
		CreateSessionKey:   auth.CreateSessionKey,
	}

	authConfig := auth.Config{
		SessionLifetime: conf.Auth.Service.SessionLifetime,
	}

	return &Manager{
		Auth: auth.NewService(s.Auth, authConfig, authTools),
	}
}
