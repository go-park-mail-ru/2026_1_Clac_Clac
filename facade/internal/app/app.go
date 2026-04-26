package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/handlers"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/router"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	"github.com/rs/zerolog"
)

type App struct {
	Config *config.Config
	Logger *zerolog.Logger
	Engine *engine.Engine
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	connector, err := NewConnector(&conf.Services)
	if err != nil {
		return nil, fmt.Errorf("NewConnector: %w", err)
	}

	manager := NewManager(connector)

	csrfSvc := usecase.NewCSRF(usecase.CSRFConfig{
		Secret: conf.CSRF.Secret,
		TTL:    24 * time.Hour,
	})

	authH := handlers.NewAuthHandler(
		manager.AuthUser,
		manager.CoolDown,
		csrfSvc,
		handlers.AuthConfig{
			MaxLenPassword:    conf.Services.Auth.Handler.MaxLenPassword,
			MinLenPassword:    conf.Services.Auth.Handler.MinLenPassword,
			SessionLifetime:   conf.Services.Auth.Handler.SessionLifetime,
			VKOAuthRedirectTo: conf.Services.Auth.Handler.VKOAuthRedirectTo,
		},
	)

	profileH := handlers.NewProfileHandler(
		manager.Profile,
		handlers.ProfileConfig{
			ValidExtensions:       config.DefaultValidExtensions(),
			SignatureTypeBytes:    conf.Services.User.Handler.SiganatureTypeBytes,
			MaxLenNameUser:        conf.Services.User.Handler.MaxLenNameUser,
			MaxLenDescriptionUser: conf.Services.User.Handler.MaxLenDescriptionUser,
			MaxReadBytes:          conf.Services.User.Handler.MaxReadBytes,
		},
	)

	r := router.NewRouter(router.RouterDeps{
		Auth:        authH,
		Profile:     profileH,
		AuthChecker: connector.Auth,
		RateLimiter: connector.RateLimiter,
		CSRFChecker: csrfSvc.Check,
	}, conf, logger)

	e := engine.New(&conf.Engine, logger, r)

	return &App{
		Config: conf,
		Logger: logger,
		Engine: e,
	}, nil
}

func (a *App) Run() {
	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func setupLogger(conf *config.Application) *zerolog.Logger {
	var out io.Writer
	if config.IsDebug(conf.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		out = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		out = os.Stdout
	}
	logger := zerolog.New(out).With().Timestamp().Logger()
	return &logger
}

