package app

import (
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
)

type Delivery struct {
	Auth       *handlers.Auth
	Profile    *handlers.Profile
	MailSender *handlers.MailSender
	CSRF       *handlers.CSRF
	Card       *handlers.Card
	Board      *handlers.Board
	Section    *handlers.Section
}

func NewDelivery(manager *Manager, conf *config.Config) *Delivery {
	authConfig := handlers.AuthConfig{
		MaxLenPassword:    conf.Services.Auth.Handler.MaxLenPassword,
		MinLenPassword:    conf.Services.Auth.Handler.MinLenPassword,
		SessionLifetime:   conf.Services.Auth.Handler.SessionLifetime,
		VKOAuthRedirectTo: conf.Services.Auth.Handler.VKOAuthRedirectTo,
	}

	profileConfig := handlers.ProfileConfig{
		ValidExtensions:       conf.Services.User.Handler.ValidExtensions,
		SignatureTypeBytes:    conf.Services.User.Handler.SignatureTypeBytes,
		MaxLenNameUser:        conf.Services.User.Handler.MaxLenNameUser,
		MaxLenDescriptionUser: conf.Services.User.Handler.MaxLenDescriptionUser,
		MaxLenPassword:        conf.Services.User.Handler.MaxLenPassword,
		MinLenPassword:        conf.Services.User.Handler.MinLenPassword,
		MaxReadBytes:          conf.Services.User.Handler.MaxReadBytes,
	}

	mailSenderConfig := handlers.MailSenderConfig{
		CoolDownExpirationSec: int64(conf.Services.RateLimiters.CoolDownExpirationSec),
	}

	cardConfig := handlers.CardConfig{
		MaxLenTitle:       conf.Services.Card.Handler.MaxLenTitle,
		MaxLenDescription: conf.Services.Card.Handler.MaxLenDescription,
	}

	boardConfig := handlers.BoardConfig{
		MultipartBackgroundFileKey: conf.Services.Board.Handler.MultipartBackgroundFileKey,
	}

	return &Delivery{
		Auth:       handlers.NewAuthHandler(manager.Auth, manager.User, authConfig),
		Profile:    handlers.NewProfileHandler(manager.User, manager.MailSender, profileConfig),
		MailSender: handlers.NewMailSender(manager.MailSender, manager.CoolDown, manager.User, mailSenderConfig),
		CSRF:       handlers.NewCSRF(manager.CSRF),
		Card:       handlers.NewCard(manager.Card, cardConfig),
		Board:      handlers.NewBoard(manager.Board, boardConfig),
		Section:    handlers.NewSection(manager.Section),
	}
}
