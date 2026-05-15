package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/delivery/http/router"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/engine"
	sentryLogger "github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/logger"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type App struct {
	Config        *config.Config
	Logger        *zerolog.Logger
	Engine        *engine.Engine
	Connector     *Connector
	MetricsServer *http.Server
}

func NewApp(conf *config.Config) (*App, error) {
	logger := setupLogger(&conf.App)

	err := setupSentry(conf)
	if err != nil {
		return nil, fmt.Errorf("setupSelery: %w", err)
	}

	metricsServer := setupMetricsServer(conf)
	go func() {
		logger.Info().Msg(fmt.Sprintf("Metrics server listening on: %s", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Metrics server failed")
		}
	}()

	connector, err := setupConnector(&conf.App, &conf.Services, logger)
	if err != nil {
		sentryLogger.CaptureError(err, "Setup connector", map[string]interface{}{"component": "connector"})
		return nil, fmt.Errorf("NewConnector: %w", err)
	}

	manager := setupManager(connector, conf)

	delivery := setupDelivery(manager, conf)

	r := setupRouter(delivery, manager, connector, conf, logger)

	e := engine.New(&conf.Engine, logger, r)

	return &App{
		Config:        conf,
		Logger:        logger,
		Engine:        e,
		Connector:     connector,
		MetricsServer: metricsServer,
	}, nil
}

func (a *App) Run() {
	defer func() {
		if a.MetricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.MetricsServer.Shutdown(ctx); err != nil {
				a.Logger.Err(err).Msg("metrics server shutdown error")
			}
		}

		a.Logger.Info().Msg("Closing gRPC connections...")

		sentryLogger.Flush()

		if a.Connector != nil {
			a.Connector.Close()
		}
	}()

	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

func setupMetricsServer(conf *config.Config) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    conf.Metrics.MetricsPort,
		Handler: mux,
	}

	return metricsServer
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

func setupConnector(app *config.Application, config *config.Services, logger *zerolog.Logger) (*Connector, error) {
	return NewConnector(app, config, logger)
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
		MailSender:  delivery.MailSender,
		CSRF:        delivery.CSRF,
		Card:        delivery.Card,
		AuthChecker: connector.Auth,
		RateLimiter: connector.RateLimiter,
		CSRFChecker: manager.CSRF.Check,
		Board:       delivery.Board,
		Section:     delivery.Section,
		Appeal:      delivery.Appeal,
	}

	return router.NewRouter(tools, conf, logger)
}

func setupSentry(config *config.Config) error {
	return sentryLogger.Init(config.Sentry)
}
