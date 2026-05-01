package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
)

type Manager struct {
	Auth       *usecase.Auth
	User       *usecase.User
	CoolDown   *usecase.CoolDown
	MailSender *usecase.MailSender
	CSRF       *usecase.CSRF
	Board      *usecase.Board
}

func NewManager(connector *Connector, conf *config.Config) *Manager {
	configCSRF := usecase.CSRFConfig{
		Secret:                         conf.CSRF.Secret,
		TTL:                            conf.CSRF.TTL,
		ExpireTimeConvertationBase:     conf.CSRF.ExpireTimeConvertationBase,
		ExpireTimeConvertationTypeSize: conf.CSRF.ExpireTimeConvertationTypeSize,
		PartsCount:                     conf.CSRF.PartsCount,
	}

	return &Manager{
		Auth:       usecase.NewAuth(connector.Auth),
		User:       usecase.NewUser(connector.User),
		CoolDown:   usecase.NewCoolDown(connector.RateLimiter),
		MailSender: usecase.NewMailSender(connector.MailSender),
		CSRF:       usecase.NewCSRF(configCSRF),
		Board:      usecase.NewBoard(connector.Board),
	}
}
