package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/config"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/engine"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/handlers"
	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func main() {
	// Настройка viper
	mainViper := viper.New()
	mainViper.SetConfigName("config")
	mainViper.SetConfigType("yaml")
	mainViper.AddConfigPath(".")

	if err := mainViper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("cannot read config file: %w", err))
	}

	// Чтение конфигов
	appConfig := config.DefaultApplicationConfig()
	if err := config.ReadWithViper(mainViper, appConfig); err != nil {
		panic(fmt.Errorf("error while setting app config: %w", err))
	}

	engineConfig := config.DefaultEngineConfig()
	if err := config.ReadWithViper(mainViper, engineConfig); err != nil {
		panic(fmt.Errorf("error while setting engine config: %w", err))
	}

	// Настройка логера
	var loggerOutput io.Writer
	if appConfig.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		loggerOutput = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		loggerOutput = os.Stdout
	}

	logger := zerolog.New(loggerOutput).With().Timestamp().Logger()

	// Добавление рутов и мидлваре
	router := mux.NewRouter()

	router.Use(middleware.RecoveryMiddleware(&logger))
	router.Use(middleware.LoggerMiddleware(&logger))

	router.HandleFunc("/healthcheck", handlers.HealthcheckHandler)

	// Создание и запуск сервера
	e := engine.New(engineConfig, &logger, router)
	if err := e.Start(context.Background()); err != nil {
		logger.Err(err).Msg("engine error")
	}
}
