package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	grpcclient "github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/clients/grpc"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/router"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
)

type App struct {
	Config  *config.Config
	Logger  *zerolog.Logger
	Engine  *engine.Engine
	Store   *Store
	Manager *Manager
	Server  *http.Server
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	clients, err := grpcclient.NewClients(&conf.Services)
	if err != nil {
		return nil, fmt.Errorf("grpcclient.NewClients: %w", err)
	}

	r := router.NewRouter(clients, conf, logger)

	srv := &http.Server{
		Addr:         conf.App.Addr,
		Handler:      r,
		ReadTimeout:  time.Duration(conf.App.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.App.WriteTimeout) * time.Second,
	}

	return &App{
		Config: conf,
		Logger: logger,
		Server: srv,
	}, nil
}

func (a *App) Run() {
	a.Logger.Info().Str("addr", a.Config.App.Addr).Msg("starting HTTP server")

	if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.Logger.Fatal().Err(err).Msg("ListenAndServe")
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
