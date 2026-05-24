package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/config"
	router "github.com/go-park-mail-ru/2026_1_Clac_Clac/realtime/internal/delivery/router"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type App struct {
	Config    config.Config
	Logger    zerolog.Logger
	Engine    *engine.Engine
	Router    *mux.Router
	Connector *Connector
	Store     *Store
	Manager   *Manager
	Delivery  *Delivery
}

func NewApp(conf config.Config) (*App, error) {
	app := &App{
		Config: conf,
	}

	app.setupLogger()

	if err := app.setupStore(); err != nil {
		return nil, fmt.Errorf("setupStore: %w", err)
	}

	if err := app.setupConnector(); err != nil {
		return nil, fmt.Errorf("setupConnector: %w", err)
	}

	app.setupManager()
	app.setupDelivery()
	app.setupRouter()
	app.setupEngine()

	return app, nil
}

func (a *App) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.Store.RedisMultiplexor.Start(ctx)

	defer func() {
		if err := a.Store.Close(); err != nil {
			a.Logger.Err(err).Msg("close store error")
		}
	}()
	defer a.Connector.Close()

	if err := a.Engine.Start(ctx); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func (a *App) setupLogger() {
	var loggerOutput io.Writer

	if config.IsDebug(a.Config.App.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	a.Logger = zerolog.New(loggerOutput).With().Timestamp().Logger()
}

func (a *App) setupStore() error {
	store, err := NewStore(&a.Logger, a.Config)
	if err != nil {
		return fmt.Errorf("store NewStore: %w", err)
	}

	a.Store = store
	return nil
}

func (a *App) setupConnector() error {
	connector, err := NewConnector(&a.Config.App, &a.Config.Services, &a.Logger)
	if err != nil {
		return fmt.Errorf("connector NewConnector: %w", err)
	}

	a.Connector = connector
	return nil
}

func (a *App) setupManager() {
	a.Manager = NewManager(a.Store, &a.Config)
}

func (a *App) setupDelivery() {
	a.Delivery = NewDelivery(a.Manager)
}

func (a *App) setupEngine() {
	a.Engine = engine.New(&a.Config.Engine, &a.Logger, a.Router)
}

func (a *App) setupRouter() {
	a.Router = router.NewRouter(router.Tools{
		Realtime:     a.Delivery.Realtime,
		AuthChecker:  a.Connector.Auth,
		BoardChecker: a.Connector.Board,
	}, &a.Config, &a.Logger)
}
