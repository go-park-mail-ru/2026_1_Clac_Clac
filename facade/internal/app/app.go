package app

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/facade/internal/engine"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type App struct {
	Config   *config.Config
	Logger   *zerolog.Logger
	Engine   *engine.Engine
	Dilivery *Delivery
	Manager  *Manager
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)

	connector := setupConnector(conf, logger)

	manager := setupManager(connector, conf)
	delivery := setupDilivery(manager, conf)

	router := setupProxyRouter( /*dilivery, manager,*/ conf, logger)

	e := setupEngine(&conf.Engine, logger, router)

	return &App{
		Config:   conf,
		Logger:   logger,
		Engine:   e,
		Delivery: delivery,
		Manager:  manager,
	}
}

// Запуск приложения
func (a *App) Run() {
	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

// Настройка рутов
func setupProxyRouter( /* dilivery *Delivery,*/ manager *Manager, conf *config.Config, logger *zerolog.Logger) *mux.Router {
	router := mux.NewRouter()

	monolitUrl, err := url.Parse(conf.MonolithURL)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(monolitUrl)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info().Msgf("[FACADE PROXY] %s %s -> %s", r.Method, r.URL.Path, conf.MonolithURL)
		proxy.ServeHTTP(w, r)
	})

	return router
}

func setupEngine(conf *config.Engine, logger *zerolog.Logger, router *mux.Router) *engine.Engine {
	return engine.New(conf, logger, router)
}

// Настройка логера
func setupLogger(conf *config.Application) *zerolog.Logger {
	var loggerOutput io.Writer

	// В зависимости от режима работы разные форматы вывода
	if config.IsDebug(conf.LogLevel) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	logger := zerolog.New(loggerOutput).With().Timestamp().Logger()
	return &logger
}

func setupConnector(conf *config.Config, logger *zerolog.Logger) *Connector {
	return NewConnector(conf, logger)
}

// Настройка менеджера сервисов
func setupManager(c *Connector, conf *config.Config) *Manager {
	return NewManager(c, conf)
}

func setupDelivery(m *Manager, conf *config.Config) *Delivery {
	return NewDelivery(m, conf)
}

// func setupVKOAuth(conf *config.VkOAuth) *oauth2.Config {
// 	return NewVKOAuth(conf)
// }
