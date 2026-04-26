package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
)

type Delivery struct {
	Auth    *handlers.Auth
	Profile *handlers.Profile
}

func NewDelivery(manager *Manager, conf *config.Config) *Delivery {
	authConfig := handlers.AuthConfig{
		MaxLenPassword:  conf.Services.Auth.Handler.MaxLenPassword,
		MinLenPassword:  conf.Services.Auth.Handler.MinLenPassword,
		SessionLifetime: conf.Services.Auth.Handler.SessionLifetime,
	}

	profileConfig := handlers.ProfileConfig{
		ValidExtensions: conf.Services.User.Handler.ValidExtensions,

		SignatureTypeBytes:    conf.Services.User.Handler.SiganatureTypeBytes,
		MaxLenNameUser:        conf.Services.User.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Services.User.Handler.MaxLenDescriptionUser,
		MaxReadBytes:          conf.Services.User.Handler.MaxReadBytes,
	}

	return &Delivery{
		Auth:    handlers.NewAuthHandler(manager.AuthUser, manager.CoolDown, manager.CSRF, authConfig),
		Profile: handlers.NewProfileHandler(manager.Profile, profileConfig),
	}
}
