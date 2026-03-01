package internal

import (
	"context"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handlers"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
)

type App struct {
	Config *config.Config
	Logger *zerolog.Logger
	Engine *engine.Engine
}

// Создает приложение, настраивает его компоненты
func NewApp(conf *config.Config) *App {
	logger := setupLogger(&conf.App)
	e := setupEngine(&conf.Engine, logger)

	return &App{
		Config: conf,
		Logger: logger,
		Engine: e,
	}
}

// Запуск приложения
func (a *App) Run() {
	if err := a.Engine.Start(context.Background()); err != nil {
		a.Logger.Err(err).Msg("engine error")
	}
}

// Настройка сервера и рутов
func setupEngine(conf *config.Engine, logger *zerolog.Logger) *engine.Engine {
	router := mux.NewRouter()

	// Добавление мидлваре
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.LoggerMiddleware(logger))

	// Добавление рутов
	router.HandleFunc("/healthcheck", handlers.HealthcheckHandler)

	// Создание сервера
	e := engine.New(conf, logger, router)
	return e
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
