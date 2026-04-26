package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
)

type Manager struct {
	AuthUser *usecase.AuthUser
	Profile  *usecase.Profile
	CoolDown *usecase.CoolDown
	CSRF     *usecase.CSRF
}

func NewManager(connector *Connector, conf *config.Config) *Manager {
	configCSRF := usecase.CSRFConfig{
		Secret: conf.CSRF.Secret,
		TTL:    conf.CSRF.TTL,
	}

	return &Manager{
		AuthUser: usecase.NewAuthUser(connector.User, connector.Auth, connector.MailSender),
		Profile:  usecase.NewProfile(connector.User),
		CoolDown: usecase.NewCoolDown(connector.RateLimiter),
		CSRF:     usecase.NewCSRF(configCSRF),
	}
}
