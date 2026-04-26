package app

import "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"

type Manager struct {
	AuthUser *usecase.AuthUser
	Profile  *usecase.Profile
	CoolDown *usecase.CoolDown
}

func NewManager(connector *Connector) *Manager {
	return &Manager{
		AuthUser: usecase.NewAuthUser(connector.User, connector.Auth, connector.MailSender),
		Profile:  usecase.NewProfile(connector.User),
		CoolDown: usecase.NewCoolDown(connector.RateLimiter),
	}
}
