package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/auth"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
)

type Manager struct {
	Auth *auth.Manager
}

func NewManager(c *Connector, conf *config.Config) *Manager {
	authManager := auth.NewManager(c.MailSenderClient)

	return &Manager{
		Auth: authManager,
	}
}
