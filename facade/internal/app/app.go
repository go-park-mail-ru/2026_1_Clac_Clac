package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/router"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type App struct {
	Config    *config.Config
	Logger    *zerolog.Logger
	Engine    *engine.Engine
	Connector *Connector
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	connector, err := setupConnector(&conf.Services, logger)
	if err != nil {
		return nil, fmt.Errorf("NewConnector: %w", err)
	}

	manager := setupManager(connector, conf)

	delivery := setupDelivery(manager, conf)

	r := setupRouter(delivery, manager, connector, conf, logger)

	e := engine.New(&conf.Engine, logger, r)

	return &App{
		Config:    conf,
		Logger:    logger,
		Engine:    e,
		Connector: connector,
	}, nil
}

func (a *App) Run() {
	defer func() {
		a.Logger.Info().Msg("Closing gRPC connections...")
		if a.Connector != nil {
			a.Connector.Close()
		}
	}()

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

func setupConnector(config *config.Services, logger *zerolog.Logger) (*Connector, error) {
	return NewConnector(config, logger)
}

func setupManager(connector *Connector, config *config.Config) *Manager {
	return NewManager(connector, config)
}

func setupDelivery(manager *Manager, conf *config.Config) *Delivery {
	return NewDelivery(manager, conf)
}

func setupRouter(delivery *Delivery, manager *Manager, connector *Connector,
	conf *config.Config, logger *zerolog.Logger) *mux.Router {
	tools := router.Tools{
		Auth:        delivery.Auth,
		Profile:     delivery.Profile,
		AuthChecker: connector.Auth,
		RateLimiter: connector.RateLimiter,
		CSRFChecker: manager.CSRF.Check,
	}

	return router.NewRouter(tools, conf, logger)
}
