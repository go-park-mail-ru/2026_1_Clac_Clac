package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
)

const recoveryEmailCooldownSec = 60

type Delivery struct {
	Auth       *handlers.Auth
	Profile    *handlers.Profile
	MailSender *handlers.MailSender
	CSRF       *handlers.CSRF
}

func NewDelivery(manager *Manager, conf *config.Config) *Delivery {
	authConfig := handlers.AuthConfig{
		MaxLenPassword:        conf.Services.Auth.Handler.MaxLenPassword,
		MinLenPassword:        conf.Services.Auth.Handler.MinLenPassword,
		SessionLifetime:       conf.Services.Auth.Handler.SessionLifetime,
		CoolDownExpirationSec: recoveryEmailCooldownSec,
	}

	profileConfig := handlers.ProfileConfig{
		ValidExtensions:       conf.Services.User.Handler.ValidExtensions,
		SignatureTypeBytes:    conf.Services.User.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.Services.User.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Services.User.Handler.MaxLenDescriptionUser,
		MaxLenPassword:        conf.Services.User.Handler.MaxLenPassword,
		MinLenPassword:        conf.Services.User.Handler.MinLenPassword,
		MaxReadBytes:          conf.Services.User.Handler.MaxReadBytes,
	}

	mailSenderConfig := handlers.MailSenderConfig{
		CoolDownExpirationSec: int64(conf.Services.RateLimiters.CoolDownExpirationSec),
	}

	return &Delivery{
		Auth:       handlers.NewAuthHandler(manager.Auth, manager.User, authConfig),
		Profile:    handlers.NewProfileHandler(manager.User, manager.MailSender, profileConfig),
		MailSender: handlers.NewMailSender(manager.MailSender, manager.CoolDown, manager.User, mailSenderConfig),
		CSRF:       handlers.NewCSRF(manager.CSRF),
	}
}
